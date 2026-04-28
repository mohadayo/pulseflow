import pytest
from app import app, events_store


@pytest.fixture
def client():
    app.config["TESTING"] = True
    events_store.clear()
    with app.test_client() as c:
        yield c


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "analytics-api"


def test_ingest_event(client):
    resp = client.post("/events", json={"type": "click", "payload": {"x": 10}})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["type"] == "click"
    assert data["payload"] == {"x": 10}
    assert "id" in data
    assert "timestamp" in data


def test_ingest_event_missing_type(client):
    resp = client.post("/events", json={"payload": {}})
    assert resp.status_code == 400
    assert "error" in resp.get_json()


def test_ingest_event_no_body(client):
    resp = client.post("/events", content_type="application/json")
    assert resp.status_code == 400


def test_list_events(client):
    client.post("/events", json={"type": "click"})
    client.post("/events", json={"type": "view"})
    resp = client.get("/events")
    assert resp.status_code == 200
    assert len(resp.get_json()) == 2


def test_list_events_filter(client):
    client.post("/events", json={"type": "click"})
    client.post("/events", json={"type": "view"})
    resp = client.get("/events?type=click")
    data = resp.get_json()
    assert len(data) == 1
    assert data[0]["type"] == "click"


def test_event_stats(client):
    client.post("/events", json={"type": "click"})
    client.post("/events", json={"type": "click"})
    client.post("/events", json={"type": "view"})
    resp = client.get("/events/stats")
    data = resp.get_json()
    assert data["total"] == 3
    assert data["by_type"]["click"] == 2
    assert data["by_type"]["view"] == 1


def test_event_stats_empty(client):
    resp = client.get("/events/stats")
    data = resp.get_json()
    assert data["total"] == 0
    assert data["by_type"] == {}
