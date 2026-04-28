const request = require("supertest");
const { app } = require("./index");

describe("API Gateway", () => {
  describe("GET /health", () => {
    it("returns ok status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("ok");
      expect(res.body.service).toBe("api-gateway");
    });
  });

  describe("POST /api/events", () => {
    it("returns 502 when upstream is unavailable", async () => {
      const res = await request(app)
        .post("/api/events")
        .send({ type: "click", payload: { x: 10 } });
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("upstream service unavailable");
    });
  });

  describe("GET /api/events", () => {
    it("returns 502 when upstream is unavailable", async () => {
      const res = await request(app).get("/api/events");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("upstream service unavailable");
    });
  });

  describe("GET /api/events/stats", () => {
    it("returns 502 when upstream is unavailable", async () => {
      const res = await request(app).get("/api/events/stats");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("upstream service unavailable");
    });
  });

  describe("GET /api/results", () => {
    it("returns 502 when upstream is unavailable", async () => {
      const res = await request(app).get("/api/results");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("upstream service unavailable");
    });
  });
});
