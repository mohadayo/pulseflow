const express = require("express");
const http = require("http");

const app = express();
app.use(express.json());

const ANALYTICS_URL = process.env.ANALYTICS_URL || "http://localhost:5001";
const PROCESSOR_URL = process.env.PROCESSOR_URL || "http://localhost:5002";
const PORT = parseInt(process.env.GATEWAY_PORT || "5000", 10);

function logger(req, _res, next) {
  const ts = new Date().toISOString();
  console.log(`${ts} ${req.method} ${req.url}`);
  next();
}

app.use(logger);

function proxyRequest(targetUrl, method, body) {
  return new Promise((resolve, reject) => {
    const url = new URL(targetUrl);
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method,
      headers: { "Content-Type": "application/json" },
    };

    const req = http.request(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          resolve({ status: res.statusCode, body: JSON.parse(data) });
        } catch {
          resolve({ status: res.statusCode, body: data });
        }
      });
    });

    req.on("error", (err) => reject(err));
    if (body) req.write(JSON.stringify(body));
    req.end();
  });
}

app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "api-gateway" });
});

app.post("/api/events", async (req, res) => {
  try {
    const analyticsResult = await proxyRequest(
      `${ANALYTICS_URL}/events`,
      "POST",
      req.body
    );

    if (analyticsResult.status !== 201) {
      return res.status(analyticsResult.status).json(analyticsResult.body);
    }

    let processorResult;
    try {
      processorResult = await proxyRequest(
        `${PROCESSOR_URL}/process`,
        "POST",
        analyticsResult.body
      );
    } catch (err) {
      console.error(`Processor unavailable: ${err.message}`);
      return res.status(201).json({
        event: analyticsResult.body,
        processing: { status: "queued", error: "processor unavailable" },
      });
    }

    res.status(201).json({
      event: analyticsResult.body,
      processing: processorResult.body,
    });
  } catch (err) {
    console.error(`Gateway error: ${err.message}`);
    res.status(502).json({ error: "upstream service unavailable" });
  }
});

app.get("/api/events", async (_req, res) => {
  try {
    const result = await proxyRequest(`${ANALYTICS_URL}/events`, "GET");
    res.status(result.status).json(result.body);
  } catch (err) {
    console.error(`Gateway error: ${err.message}`);
    res.status(502).json({ error: "upstream service unavailable" });
  }
});

app.get("/api/events/stats", async (_req, res) => {
  try {
    const result = await proxyRequest(
      `${ANALYTICS_URL}/events/stats`,
      "GET"
    );
    res.status(result.status).json(result.body);
  } catch (err) {
    console.error(`Gateway error: ${err.message}`);
    res.status(502).json({ error: "upstream service unavailable" });
  }
});

app.get("/api/results", async (_req, res) => {
  try {
    const result = await proxyRequest(`${PROCESSOR_URL}/results`, "GET");
    res.status(result.status).json(result.body);
  } catch (err) {
    console.error(`Gateway error: ${err.message}`);
    res.status(502).json({ error: "upstream service unavailable" });
  }
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`API Gateway listening on port ${PORT}`);
  });
}

module.exports = { app, proxyRequest };
