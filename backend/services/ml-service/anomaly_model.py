from sklearn.ensemble import IsolationForest
import numpy as np

class AnomalyModel:
    def __init__(self, contamination=0.05):
        self.model = IsolationForest(
            contamination=contamination,
            random_state=42,
            n_estimators=100
        )
        self.is_trained = False

    def train(self, X):
        """Train anomaly detection model"""
        self.model.fit(X)
        self.is_trained = True

    def detect(self, X):
        """Detect anomalies (-1 = anomaly, 1 = normal)"""
        if not self.is_trained:
            raise RuntimeError("Model not trained")

        predictions = self.model.predict(X)
        scores = self.model.score_samples(X)

        return predictions, scores

    def get_baseline_stats(self, X):
        """Calculate baseline statistics for z-score detection"""
        return {
            "mean": float(np.mean(X)),
            "std": float(np.std(X)),
            "min": float(np.min(X)),
            "max": float(np.max(X))
        }
