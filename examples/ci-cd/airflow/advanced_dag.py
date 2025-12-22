"""
Advanced Airflow DAG with Error Handling and Quality Checks

Features:
- Multi-step ETL pipeline
- Git repository cloning
- Error handling and notifications
- Quality checks
- JSON report parsing

Usage:
1. Copy this file to your Airflow DAGs folder
2. Set Airflow Variables: VECTOR_DB_TYPE, QDRANT_API_KEY, QDRANT_URL, OPENAI_API_KEY
3. Configure email alerts in Airflow
"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.operators.python import PythonOperator, BranchPythonOperator
from airflow.operators.email import EmailOperator
from airflow.models import Variable
import json

default_args = {
    'owner': 'data-team',
    'depends_on_past': False,
    'start_date': datetime(2025, 1, 1),
    'email': ['team@example.com'],
    'email_on_failure': True,
    'email_on_retry': False,
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG(
    'weave_advanced_etl',
    default_args=default_args,
    description='Advanced ETL pipeline for documentation ingestion',
    schedule_interval='0 2 * * *',
    catchup=False,
    tags=['weave', 'etl', 'vector-db'],
)

# Environment variables
env_vars = {
    'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
    'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
    'QDRANT_URL': Variable.get('QDRANT_URL'),
    'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
}


def parse_batch_report(**context):
    """Parse batch report and determine next action"""
    with open('/tmp/batch-report.json', 'r') as f:
        report = json.load(f)

    # Push metrics to XCom
    context['task_instance'].xcom_push(key='files_processed', value=report['files']['processed'])
    context['task_instance'].xcom_push(key='files_failed', value=report['files']['failed'])
    context['task_instance'].xcom_push(key='exit_code', value=report['exit_code'])

    # Determine branch based on exit code
    exit_code = report['exit_code']
    if exit_code == 0:
        return 'success_notification'
    elif exit_code == 1:
        return 'partial_failure_notification'
    else:
        return 'complete_failure_notification'


def quality_check(**context):
    """Verify ingestion quality"""
    ti = context['task_instance']
    files_processed = ti.xcom_pull(task_ids='parse_report', key='files_processed')
    files_failed = ti.xcom_pull(task_ids='parse_report', key='files_failed')

    if files_processed == 0:
        raise ValueError("No files were processed!")

    failure_rate = files_failed / (files_processed + files_failed)
    if failure_rate > 0.1:  # More than 10% failure rate
        raise ValueError(f"Failure rate too high: {failure_rate:.1%}")

    print(f"✅ Quality check passed. Processed: {files_processed}, Failed: {files_failed}")


# Task 1: Clone repository
clone_repo = BashOperator(
    task_id='clone_repository',
    bash_command='''
    cd /opt/airflow/data
    if [ -d "docs-repo" ]; then
        cd docs-repo && git pull
        echo "Repository updated"
    else
        git clone https://github.com/your-org/docs-repo.git
        echo "Repository cloned"
    fi
    ''',
    dag=dag,
)

# Task 2: Ingest documents
ingest = BashOperator(
    task_id='ingest_documents',
    bash_command='''
    weave docs batch \
      --directory /opt/airflow/data/docs-repo/docs \
      --collection documentation \
      --parallel 3 \
      --json > /tmp/batch-report.json

    EXIT_CODE=$?
    cat /tmp/batch-report.json
    echo "Exit code: $EXIT_CODE"

    # Allow partial failures to continue pipeline
    if [ $EXIT_CODE -eq 2 ]; then
        exit 2
    fi
    exit 0
    ''',
    env=env_vars,
    dag=dag,
)

# Task 3: Parse report
parse_report = PythonOperator(
    task_id='parse_report',
    python_callable=parse_batch_report,
    provide_context=True,
    dag=dag,
)

# Task 4: Branch based on results
branch = BranchPythonOperator(
    task_id='check_result',
    python_callable=parse_batch_report,
    provide_context=True,
    dag=dag,
)

# Task 5a: Success notification
success = EmailOperator(
    task_id='success_notification',
    to='team@example.com',
    subject='✅ Weave Ingestion Successful',
    html_content='''
    <h3>Document Ingestion Completed Successfully</h3>
    <p><strong>Files processed:</strong> {{ task_instance.xcom_pull(task_ids='parse_report', key='files_processed') }}</p>
    ''',
    dag=dag,
)

# Task 5b: Partial failure notification
partial_failure = EmailOperator(
    task_id='partial_failure_notification',
    to='team@example.com',
    subject='⚠️ Weave Ingestion Partial Failure',
    html_content='''
    <h3>Document Ingestion Completed with Some Failures</h3>
    <p><strong>Files processed:</strong> {{ task_instance.xcom_pull(task_ids='parse_report', key='files_processed') }}</p>
    <p><strong>Files failed:</strong> {{ task_instance.xcom_pull(task_ids='parse_report', key='files_failed') }}</p>
    ''',
    dag=dag,
)

# Task 5c: Complete failure notification
complete_failure = EmailOperator(
    task_id='complete_failure_notification',
    to='team@example.com',
    subject='❌ Weave Ingestion Failed',
    html_content='<h3>Document Ingestion Failed Completely</h3>',
    dag=dag,
)

# Task 6: Quality check
quality = PythonOperator(
    task_id='quality_check',
    python_callable=quality_check,
    provide_context=True,
    trigger_rule='none_failed',
    dag=dag,
)

# Define dependencies
clone_repo >> ingest >> parse_report >> branch
branch >> [success, partial_failure, complete_failure] >> quality
