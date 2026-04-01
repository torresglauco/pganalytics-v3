from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import numpy as np
from models import LatencyModel
import logging

app = FastAPI(title="pgAnalytics ML Service")
logger = logging.getLogger(__name__)

# Load pre-trained model
latency_model = LatencyModel()
try:
    latency_model.load("./models")
except Exception as e:
    logger.warning(f"Could not load pre-trained model: {e}")

class PredictionRequest(BaseModel):
    features: list[float]

class PredictionResponse(BaseModel):
    latency_ms: float
    confidence: float

@app.post("/predict")
def predict_latency(request: PredictionRequest) -> PredictionResponse:
    """Predict query latency based on features"""
    try:
        X = np.array([request.features])
        pred = latency_model.predict(X)[0]

        # Confidence based on model certainty (simplified)
        confidence = 0.85

        return PredictionResponse(
            latency_ms=float(pred),
            confidence=confidence
        )
    except Exception as e:
        logger.error(f"Prediction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5000)
