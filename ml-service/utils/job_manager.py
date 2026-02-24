"""
Job management utilities for tracking async tasks
"""

import logging
from typing import Dict, Any, Optional
from datetime import datetime
import uuid

logger = logging.getLogger(__name__)


class JobStatus:
    """Job status constants"""
    PENDING = 'pending'
    TRAINING = 'training'
    COMPLETED = 'completed'
    FAILED = 'failed'


class JobManager:
    """
    Manages async job tracking and status updates

    In a full implementation, this would use Redis or database for persistence.
    For now, uses in-memory storage for testing.
    """

    # In-memory job store (in production, use Redis or database)
    _jobs: Dict[str, Dict[str, Any]] = {}

    @staticmethod
    def generate_job_id(prefix: str = 'job') -> str:
        """
        Generate unique job identifier

        Args:
            prefix: Prefix for job ID

        Returns:
            Unique job ID
        """
        timestamp = datetime.utcnow().strftime('%Y%m%d-%H%M%S')
        random_part = str(uuid.uuid4())[:8]
        return f"{prefix}-{timestamp}-{random_part}"

    @staticmethod
    def create_job(job_type: str, **kwargs) -> Dict[str, Any]:
        """
        Create a new job record

        Args:
            job_type: Type of job (training, validation, etc)
            **kwargs: Additional job parameters

        Returns:
            Job record with ID and metadata
        """
        job_id = JobManager.generate_job_id(prefix=job_type)

        job = {
            'job_id': job_id,
            'job_type': job_type,
            'status': JobStatus.PENDING,
            'created_at': datetime.utcnow().isoformat(),
            'updated_at': datetime.utcnow().isoformat(),
            'result': None,
            'error': None,
        }

        # Add additional parameters
        job.update(kwargs)

        # Store job
        JobManager._jobs[job_id] = job

        logger.info(f"Job created: {job_id} (type={job_type})")
        return job

    @staticmethod
    def get_job(job_id: str) -> Optional[Dict[str, Any]]:
        """
        Get job by ID

        Args:
            job_id: Job identifier

        Returns:
            Job record or None if not found
        """
        return JobManager._jobs.get(job_id)

    @staticmethod
    def update_job(job_id: str, **kwargs) -> Optional[Dict[str, Any]]:
        """
        Update job record

        Args:
            job_id: Job identifier
            **kwargs: Fields to update

        Returns:
            Updated job record or None if not found
        """
        job = JobManager._jobs.get(job_id)
        if not job:
            logger.warning(f"Job not found: {job_id}")
            return None

        job.update(kwargs)
        job['updated_at'] = datetime.utcnow().isoformat()

        logger.debug(f"Job updated: {job_id}")
        return job

    @staticmethod
    def set_status(job_id: str, status: str) -> Optional[Dict[str, Any]]:
        """
        Set job status

        Args:
            job_id: Job identifier
            status: New status

        Returns:
            Updated job record
        """
        return JobManager.update_job(job_id, status=status)

    @staticmethod
    def set_result(job_id: str, result: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """
        Set job result

        Args:
            job_id: Job identifier
            result: Result dictionary

        Returns:
            Updated job record
        """
        return JobManager.update_job(
            job_id,
            status=JobStatus.COMPLETED,
            result=result,
            completed_at=datetime.utcnow().isoformat()
        )

    @staticmethod
    def set_error(job_id: str, error: str) -> Optional[Dict[str, Any]]:
        """
        Set job error

        Args:
            job_id: Job identifier
            error: Error message

        Returns:
            Updated job record
        """
        return JobManager.update_job(
            job_id,
            status=JobStatus.FAILED,
            error=error,
            failed_at=datetime.utcnow().isoformat()
        )

    @staticmethod
    def list_jobs(job_type: Optional[str] = None, status: Optional[str] = None) -> list:
        """
        List jobs with optional filtering

        Args:
            job_type: Filter by job type (optional)
            status: Filter by status (optional)

        Returns:
            List of job records
        """
        jobs = list(JobManager._jobs.values())

        if job_type:
            jobs = [j for j in jobs if j.get('job_type') == job_type]

        if status:
            jobs = [j for j in jobs if j.get('status') == status]

        return jobs

    @staticmethod
    def clear_old_jobs(max_age_hours: int = 24) -> int:
        """
        Clear jobs older than specified age

        Args:
            max_age_hours: Maximum age in hours

        Returns:
            Number of jobs deleted
        """
        from datetime import timedelta

        cutoff_time = datetime.utcnow() - timedelta(hours=max_age_hours)
        count = 0

        job_ids_to_delete = []
        for job_id, job in JobManager._jobs.items():
            try:
                created_at = datetime.fromisoformat(job['created_at'])
                if created_at < cutoff_time:
                    job_ids_to_delete.append(job_id)
            except (ValueError, KeyError):
                pass

        for job_id in job_ids_to_delete:
            del JobManager._jobs[job_id]
            count += 1

        logger.info(f"Cleaned up {count} old jobs")
        return count


class TrainingJobManager:
    """Helper class for managing training jobs specifically"""

    @staticmethod
    def create_training_job(
        database_name: str,
        lookback_days: int,
        model_type: str,
        force_retrain: bool = False
    ) -> Dict[str, Any]:
        """
        Create a training job

        Args:
            database_name: Name of database
            lookback_days: Days of historical data
            model_type: Model type to train
            force_retrain: Force retraining if model exists

        Returns:
            Job record
        """
        return JobManager.create_job(
            job_type='training',
            database_name=database_name,
            lookback_days=lookback_days,
            model_type=model_type,
            force_retrain=force_retrain,
        )

    @staticmethod
    def get_training_job(job_id: str) -> Optional[Dict[str, Any]]:
        """Get training job by ID"""
        job = JobManager.get_job(job_id)
        if job and job.get('job_type') == 'training':
            return job
        return None

    @staticmethod
    def mark_training_started(job_id: str) -> Optional[Dict[str, Any]]:
        """Mark training as started"""
        return JobManager.set_status(job_id, JobStatus.TRAINING)

    @staticmethod
    def mark_training_completed(job_id: str, model_id: str, metrics: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Mark training as completed with results"""
        result = {
            'model_id': model_id,
            **metrics
        }
        return JobManager.set_result(job_id, result)

    @staticmethod
    def mark_training_failed(job_id: str, error: str) -> Optional[Dict[str, Any]]:
        """Mark training as failed"""
        return JobManager.set_error(job_id, error)
