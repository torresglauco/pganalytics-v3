"""
Celery worker entry point for running background tasks
"""

import logging
import sys
from tasks import celery_app

logger = logging.getLogger(__name__)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)


def run_worker():
    """Run Celery worker"""
    try:
        logger.info("Starting Celery worker...")

        celery_app.worker_main([
            'worker',
            '--loglevel=info',
            '--concurrency=4',
            '--pool=threads',
        ])
    except KeyboardInterrupt:
        logger.info("Worker shutdown requested")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Error starting worker: {e}")
        sys.exit(1)


if __name__ == '__main__':
    run_worker()
