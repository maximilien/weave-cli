"""
Incremental Update DAG

This DAG runs hourly and processes only files modified in the last hour.
Ideal for continuous ingestion with minimal resource usage.

Usage:
1. Copy this file to your Airflow DAGs folder
2. Set Airflow Variables: VECTOR_DB_TYPE, QDRANT_API_KEY, QDRANT_URL, OPENAI_API_KEY
3. Adjust schedule_interval as needed
"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.models import Variable

default_args = {
    'owner': 'data-team',
    'depends_on_past': False,
    'start_date': datetime(2025, 1, 1),
    'email': ['alerts@example.com'],
    'email_on_failure': True,
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG(
    'weave_incremental_update',
    default_args=default_args,
    description='Hourly incremental documentation updates',
    schedule_interval='0 * * * *',  # Every hour
    catchup=False,
    max_active_runs=1,  # Prevent overlapping runs
    tags=['weave', 'incremental', 'vector-db'],
)

# Environment variables
env_vars = {
    'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE', default_var='qdrant-cloud'),
    'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
    'QDRANT_URL': Variable.get('QDRANT_URL'),
    'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
}

# Task: Incremental ingestion (only files modified in last hour)
incremental_ingest = BashOperator(
    task_id='incremental_ingest',
    bash_command='''
    weave docs batch \
      --directory /opt/airflow/data/docs \
      --collection documentation \
      --since 1h \
      --parallel 2 \
      --json > /tmp/incremental-report.json

    EXIT_CODE=$?

    # Parse and log results
    FILES_PROCESSED=$(jq -r '.files.processed' /tmp/incremental-report.json)
    DOCS_CREATED=$(jq -r '.documents.created' /tmp/incremental-report.json)

    echo "📊 Incremental Update Results:"
    echo "  Files processed: $FILES_PROCESSED"
    echo "  Documents created: $DOCS_CREATED"

    cat /tmp/incremental-report.json
    exit $EXIT_CODE
    ''',
    env=env_vars,
    dag=dag,
)
