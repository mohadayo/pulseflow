import logging
import os
import time
import uuid
from flask import Flask, jsonify, request

app = Flask(__name__)

LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO").upper()
logging.basicConfig(level=LOG_LEVEL, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

events_store: list[dict] = []


@app.get("/health")
def health():
    return jsonify({"status": "ok", "service": "analytics-api"})


@app.post("/events")
def ingest_event():
    body = request.get_json(silent=True)
    if not body or "type" not in body:
        logger.warning("Rejected event: missing 'type' field")
        return jsonify({"error": "field 'type' is required"}), 400

    event = {
        "id": str(uuid.uuid4()),
        "type": body["type"],
        "payload": body.get("payload", {}),
        "timestamp": time.time(),
    }
    events_store.append(event)
    logger.info("Ingested event %s of type %s", event["id"], event["type"])
    return jsonify(event), 201


@app.get("/events")
def list_events():
    event_type = request.args.get("type")
    results = events_store
    if event_type:
        results = [e for e in events_store if e["type"] == event_type]
    logger.info("Listing %d events (filter=%s)", len(results), event_type)
    return jsonify(results)


@app.get("/events/stats")
def event_stats():
    type_counts: dict[str, int] = {}
    for e in events_store:
        type_counts[e["type"]] = type_counts.get(e["type"], 0) + 1
    stats = {"total": len(events_store), "by_type": type_counts}
    logger.info("Stats requested: %d total events", stats["total"])
    return jsonify(stats)


def create_app():
    return app


if __name__ == "__main__":
    port = int(os.environ.get("ANALYTICS_PORT", "5001"))
    app.run(host="0.0.0.0", port=port)
