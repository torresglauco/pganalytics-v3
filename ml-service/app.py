"""
ML Service Flask Application
Provides machine learning endpoints for query performance prediction
"""

import logging
import os
from flask import Flask, jsonify
from flask_cors import CORS
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def create_app(config_name=None):
    """
    Application factory function

    Args:
        config_name: 'development', 'production', 'testing'
    """
    from config import DevelopmentConfig, ProductionConfig, TestingConfig

    app = Flask(__name__)

    # Select configuration
    if config_name is None:
        config_name = os.getenv('FLASK_ENV', 'development')

    config_map = {
        'development': DevelopmentConfig,
        'production': ProductionConfig,
        'testing': TestingConfig,
    }

    config = config_map.get(config_name, DevelopmentConfig)
    app.config.from_object(config)

    logger.info(f"Initialized Flask app with {config_name} configuration")

    # Enable CORS
    CORS(app, resources={r"/api/*": {"origins": app.config['CORS_ORIGINS']}})
    logger.info("CORS enabled")

    # Register blueprint for API routes
    from api.routes import api_blueprint
    app.register_blueprint(api_blueprint, url_prefix='/api')

    # Health check endpoint
    @app.route('/health', methods=['GET'])
    def health_check():
        """Health check endpoint for container orchestration"""
        return jsonify({'status': 'healthy', 'service': 'ml-service'}), 200

    # Error handlers
    @app.errorhandler(400)
    def bad_request(error):
        """Handle 400 Bad Request errors"""
        logger.warning(f"Bad request: {error}")
        return jsonify({'error': 'Bad request', 'message': str(error)}), 400

    @app.errorhandler(404)
    def not_found(error):
        """Handle 404 Not Found errors"""
        logger.warning(f"Not found: {error}")
        return jsonify({'error': 'Not found', 'message': 'Endpoint not found'}), 404

    @app.errorhandler(500)
    def internal_error(error):
        """Handle 500 Internal Server Error"""
        logger.error(f"Internal server error: {error}")
        return jsonify({'error': 'Internal server error', 'message': 'An unexpected error occurred'}), 500

    logger.info("Flask application created successfully")
    return app


if __name__ == '__main__':
    app = create_app()
    port = int(os.getenv('API_PORT', 8081))
    debug = os.getenv('FLASK_ENV', 'development') == 'development'

    logger.info(f"Starting ML Service on port {port}")
    app.run(
        host='0.0.0.0',
        port=port,
        debug=debug,
        use_reloader=False  # Disable for production
    )
