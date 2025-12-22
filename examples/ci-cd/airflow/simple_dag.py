"""
Simple Airflow DAG for Document Ingestion

This DAG runs daily at 2 AM and ingests documentation into a vector database.

Usage:
1. Copy this file to your Airflow DAGs folder
2. Set Airflow Variables: VECTOR_DB_TYPE, QDRANT_API_KEY, QDRANT_URL, OPENAI_API_KEY
3. The DAG will appear in the Airflow UI and can be triggered manually or run on schedule
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
    'email_on_retry': False,
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG(
    'weave_ingest_documentation',
    default_args=default_args,
    description='Ingest documentation into vector database',
    schedule_interval='0 2 * * *',  # Daily at 2 AM
    catchup=False,
    tags=['weave', 'documentation', 'vector-db'],
)

# Environment variables from Airflow Variables
env_vars = {
    'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE', default_var='qdrant-cloud'),
    'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
    'QDRANT_URL': Variable.get('QDRANT_URL'),
    'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
    'EMBEDDING_MODEL': Variable.get('EMBEDDING_MODEL', default_var='text-embedding-3-small'),
}

# Task: Ingest documents
ingest_documents = BashOperator(
    task_id='ingest_documents',
    bash_command='''
    weave docs batch \
      --directory /opt/airflow/data/docs \
      --collection documentation \
      --parallel 3 \
      --json > /tmp/batch-report.json

    EXIT_CODE=$?
    echo "Exit code: $EXIT_CODE"
    cat /tmp/batch-report.json
    exit $EXIT_CODE
    ''',
    env=env_vars,
    dag=dag,
)
