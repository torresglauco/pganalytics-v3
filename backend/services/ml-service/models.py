import numpy as np
import joblib
from sklearn.ensemble import RandomForestRegressor
from sklearn.preprocessing import StandardScaler

class LatencyModel:
    def __init__(self):
        self.model = None
        self.scaler = StandardScaler()
        self.is_trained = False

    def train(self, X, y):
        """Train latency prediction model"""
        X_scaled = self.scaler.fit_transform(X)
        self.model = RandomForestRegressor(
            n_estimators=100,
            max_depth=15,
            min_samples_split=5,
            random_state=42,
            n_jobs=-1
        )
        self.model.fit(X_scaled, y)
        self.is_trained = True

    def predict(self, X):
        """Predict latency for query features"""
        if not self.is_trained:
            raise RuntimeError("Model not trained")

        X_scaled = self.scaler.transform(X)
        return self.model.predict(X_scaled)

    def save(self, path):
        """Save model to disk"""
        joblib.dump(self.model, f"{path}/model.joblib")
        joblib.dump(self.scaler, f"{path}/scaler.joblib")

    def load(self, path):
        """Load model from disk"""
        self.model = joblib.load(f"{path}/model.joblib")
        self.scaler = joblib.load(f"{path}/scaler.joblib")
        self.is_trained = True
