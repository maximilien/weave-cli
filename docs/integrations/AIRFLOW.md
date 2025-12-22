# Apache Airflow Integration Guide

This guide shows how to use `weave docs batch` in Apache Airflow DAGs for automated document ingestion into vector databases.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Basic DAG](#basic-dag)
- [DAG Examples](#dag-examples)
- [Operators](#operators)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

Apache Airflow is a platform to programmatically author, schedule, and monitor workflows. It's ideal for:
- Scheduled document ingestion
- Complex data pipelines
- Dependency management
- Retry and failure handling

## Prerequisites

1. Apache Airflow 2.0+ installed
2. weave-cli available in Airflow environment
3. VDB credentials configured in Airflow Connections or Variables
4. Python 3.8+

### Install weave-cli in Airflow Environment

**Option 1: System Installation**

```bash
# In Airflow container/environment
curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o /usr/local/bin/weave
chmod +x /usr/local/bin/weave
```

**Option 2: Custom Docker Image**

```dockerfile
FROM apache/airflow:2.7.0

USER root
RUN curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o /usr/local/bin/weave && \
    chmod +x /usr/local/bin/weave

USER airflow
```

## Basic DAG

### Simple Document Ingestion DAG

```python
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.operators.python import PythonOperator
from airflow.models import Variable
import json

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

# Set environment variables from Airflow Variables
env_vars = {
    'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE', default_var='qdrant-cloud'),
    'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
    'QDRANT_URL': Variable.get('QDRANT_URL'),
    'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
    'EMBEDDING_MODEL': Variable.get('EMBEDDING_MODEL', default_var='text-embedding-3-small'),
}

ingest_task = BashOperator(
    task_id='ingest_documents',
    bash_command='''
    weave docs batch \
      --directory /opt/airflow/data/docs \
      --collection documentation \
      --parallel 3 \
      --json > /tmp/batch-report.json

    EXIT_CODE=$?
    cat /tmp/batch-report.json
    exit $EXIT_CODE
    ''',
    env=env_vars,
    dag=dag,
)
```

## DAG Examples

### Example 1: Complete ETL Pipeline with Quality Checks

```python
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.operators.python import PythonOperator, BranchPythonOperator
from airflow.operators.email import EmailOperator
from airflow.models import Variable
from datetime import datetime, timedelta
import json

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
        return 'partial_failure_handler'
    else:
        return 'complete_failure_handler'

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

    print(f"Quality check passed. Processed: {files_processed}, Failed: {files_failed}")

with DAG(
    'weave_etl_pipeline',
    default_args={
        'owner': 'data-team',
        'retries': 3,
        'retry_delay': timedelta(minutes=5),
        'start_date': datetime(2025, 1, 1),
    },
    schedule_interval='0 2 * * *',
    catchup=False,
    tags=['weave', 'etl', 'vector-db'],
) as dag:

    # Step 1: Clone/update documentation repository
    clone_repo = BashOperator(
        task_id='clone_repository',
        bash_command='''
        cd /opt/airflow/data
        if [ -d "docs-repo" ]; then
            cd docs-repo && git pull
        else
            git clone https://github.com/your-org/docs-repo.git
        fi
        '''
    )

    # Step 2: Run weave batch ingestion
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
        env={
            'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
            'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
            'QDRANT_URL': Variable.get('QDRANT_URL'),
            'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
        },
    )

    # Step 3: Parse batch report
    parse_report = PythonOperator(
        task_id='parse_report',
        python_callable=parse_batch_report,
        provide_context=True,
    )

    # Step 4: Branch based on results
    branch = BranchPythonOperator(
        task_id='check_result',
        python_callable=parse_batch_report,
        provide_context=True,
    )

    # Step 5a: Handle success
    success = EmailOperator(
        task_id='success_notification',
        to='team@example.com',
        subject='Weave Ingestion Successful',
        html_content='''
        <h3>Document Ingestion Completed Successfully</h3>
        <p>Files processed: {{ task_instance.xcom_pull(task_ids='parse_report', key='files_processed') }}</p>
        ''',
    )

    # Step 5b: Handle partial failure
    partial_failure = EmailOperator(
        task_id='partial_failure_handler',
        to='team@example.com',
        subject='Weave Ingestion Partial Failure',
        html_content='''
        <h3>Document Ingestion Completed with Some Failures</h3>
        <p>Files processed: {{ task_instance.xcom_pull(task_ids='parse_report', key='files_processed') }}</p>
        <p>Files failed: {{ task_instance.xcom_pull(task_ids='parse_report', key='files_failed') }}</p>
        ''',
    )

    # Step 5c: Handle complete failure
    complete_failure = EmailOperator(
        task_id='complete_failure_handler',
        to='team@example.com',
        subject='Weave Ingestion Failed',
        html_content='<h3>Document Ingestion Failed Completely</h3>',
    )

    # Step 6: Quality check
    quality = PythonOperator(
        task_id='quality_check',
        python_callable=quality_check,
        provide_context=True,
        trigger_rule='none_failed',
    )

    # Define dependencies
    clone_repo >> ingest >> parse_report >> branch
    branch >> [success, partial_failure, complete_failure] >> quality
```

### Example 2: Incremental Daily Updates

```python
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.sensors.filesystem import FileSensor
from datetime import datetime, timedelta

with DAG(
    'weave_incremental_update',
    default_args={
        'owner': 'data-team',
        'retries': 3,
        'start_date': datetime(2025, 1, 1),
    },
    schedule_interval='0 * * * *',  # Hourly
    catchup=False,
    tags=['weave', 'incremental'],
) as dag:

    # Process only files modified in last hour
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

        # Log results
        FILES_PROCESSED=$(jq -r '.files.processed' /tmp/incremental-report.json)
        echo "Processed $FILES_PROCESSED files in last hour"

        cat /tmp/incremental-report.json
        exit $EXIT_CODE
        ''',
        env={
            'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
            'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
            'QDRANT_URL': Variable.get('QDRANT_URL'),
            'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
        },
    )
```

### Example 3: Multi-Collection DAG with Dynamic Tasks

```python
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.operators.python import PythonOperator
from airflow.models import Variable
from datetime import datetime, timedelta

# Define collections to process
COLLECTIONS = [
    {'directory': '/opt/airflow/data/docs/api', 'collection': 'api-docs'},
    {'directory': '/opt/airflow/data/docs/guides', 'collection': 'guides'},
    {'directory': '/opt/airflow/data/docs/tutorials', 'collection': 'tutorials'},
    {'directory': '/opt/airflow/data/docs/reference', 'collection': 'reference'},
]

def generate_tasks(dag, collections):
    """Dynamically generate tasks for each collection"""
    tasks = []

    for config in collections:
        task = BashOperator(
            task_id=f"ingest_{config['collection']}",
            bash_command=f'''
            weave docs batch \
              --directory {config['directory']} \
              --collection {config['collection']} \
              --parallel 2 \
              --json > /tmp/{config['collection']}-report.json

            EXIT_CODE=$?
            cat /tmp/{config['collection']}-report.json
            exit $EXIT_CODE
            ''',
            env={
                'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
                'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
                'QDRANT_URL': Variable.get('QDRANT_URL'),
                'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
            },
            dag=dag,
        )
        tasks.append(task)

    return tasks

with DAG(
    'weave_multi_collection',
    default_args={
        'owner': 'data-team',
        'retries': 2,
        'start_date': datetime(2025, 1, 1),
    },
    schedule_interval='0 2 * * *',
    catchup=False,
    max_active_runs=1,
    tags=['weave', 'multi-collection'],
) as dag:

    start = BashOperator(
        task_id='start',
        bash_command='echo "Starting multi-collection ingestion"',
    )

    # Generate tasks dynamically
    collection_tasks = generate_tasks(dag, COLLECTIONS)

    # Summary task
    summary = PythonOperator(
        task_id='create_summary',
        python_callable=lambda: print("All collections processed!"),
    )

    # Set dependencies
    start >> collection_tasks >> summary
```

### Example 4: DAG with Sensor for File Availability

```python
from airflow import DAG
from airflow.sensors.filesystem import FileSensor
from airflow.operators.bash import BashOperator
from datetime import datetime, timedelta

with DAG(
    'weave_sensor_based',
    default_args={
        'owner': 'data-team',
        'retries': 3,
        'start_date': datetime(2025, 1, 1),
    },
    schedule_interval='@daily',
    catchup=False,
    tags=['weave', 'sensor'],
) as dag:

    # Wait for new documents to arrive
    wait_for_docs = FileSensor(
        task_id='wait_for_documents',
        filepath='/opt/airflow/data/docs/*.md',
        poke_interval=300,  # Check every 5 minutes
        timeout=3600,  # Timeout after 1 hour
        mode='poke',
    )

    # Process documents when available
    ingest = BashOperator(
        task_id='ingest_documents',
        bash_command='''
        weave docs batch \
          --directory /opt/airflow/data/docs \
          --collection documentation \
          --parallel 3 \
          --json > /tmp/batch-report.json
        ''',
        env={
            'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
            'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
            'QDRANT_URL': Variable.get('QDRANT_URL'),
            'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
        },
    )

    wait_for_docs >> ingest
```

## Operators

### Custom WeaveOperator

Create a reusable operator for weave-cli:

```python
from airflow.models import BaseOperator
from airflow.utils.decorators import apply_defaults
from airflow.models import Variable
import subprocess
import json

class WeaveIngestOperator(BaseOperator):
    """
    Operator for running weave docs batch command

    :param directory: Directory containing documents
    :param collection: Collection name
    :param parallel: Number of parallel workers
    :param since: Only process files modified since duration
    :param vdb_type: Vector database type
    """

    template_fields = ('directory', 'collection')

    @apply_defaults
    def __init__(
        self,
        directory,
        collection,
        parallel=3,
        since=None,
        vdb_type=None,
        *args,
        **kwargs
    ):
        super().__init__(*args, **kwargs)
        self.directory = directory
        self.collection = collection
        self.parallel = parallel
        self.since = since
        self.vdb_type = vdb_type or Variable.get('VECTOR_DB_TYPE')

    def execute(self, context):
        # Build command
        cmd = [
            'weave', 'docs', 'batch',
            '--directory', self.directory,
            '--collection', self.collection,
            '--parallel', str(self.parallel),
            '--json',
        ]

        if self.since:
            cmd.extend(['--since', self.since])

        # Set environment
        env = {
            'VECTOR_DB_TYPE': self.vdb_type,
            'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
            'QDRANT_URL': Variable.get('QDRANT_URL'),
            'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
        }

        # Execute command
        self.log.info(f"Executing: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            env=env,
            capture_output=True,
            text=True
        )

        # Parse JSON output
        try:
            report = json.loads(result.stdout)
            self.log.info(f"Batch report: {json.dumps(report, indent=2)}")

            # Push metrics to XCom
            context['task_instance'].xcom_push(key='batch_report', value=report)
            context['task_instance'].xcom_push(key='files_processed', value=report['files']['processed'])
            context['task_instance'].xcom_push(key='files_failed', value=report['files']['failed'])

            # Check exit code
            if report['exit_code'] == 2:
                raise Exception(f"Batch ingestion failed completely")
            elif report['exit_code'] == 1:
                self.log.warning("Batch ingestion completed with some failures")

        except json.JSONDecodeError:
            self.log.error(f"Failed to parse JSON output: {result.stdout}")
            raise

        return report

# Usage in DAG
with DAG('weave_custom_operator', ...) as dag:
    ingest = WeaveIngestOperator(
        task_id='ingest_docs',
        directory='/opt/airflow/data/docs',
        collection='documentation',
        parallel=3,
        since='24h',
    )
```

## Error Handling

### Retry Configuration

```python
default_args = {
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
    'retry_exponential_backoff': True,
    'max_retry_delay': timedelta(minutes=30),
}
```

### On-Failure Callback

```python
def on_failure_callback(context):
    """Send notification on task failure"""
    task_instance = context['task_instance']
    exception = context.get('exception')

    # Log error
    print(f"Task {task_instance.task_id} failed: {exception}")

    # Send alert (Slack, PagerDuty, etc.)
    # send_slack_alert(...)

default_args = {
    'on_failure_callback': on_failure_callback,
}
```

## Best Practices

### 1. Use Airflow Variables for Configuration

Store configuration in Airflow Variables instead of hardcoding:

```python
env_vars = {
    'VECTOR_DB_TYPE': Variable.get('VECTOR_DB_TYPE'),
    'QDRANT_API_KEY': Variable.get('QDRANT_API_KEY'),
    'QDRANT_URL': Variable.get('QDRANT_URL'),
    'OPENAI_API_KEY': Variable.get('OPENAI_API_KEY'),
}
```

### 2. Use XCom for Sharing Data

Share batch reports between tasks:

```python
# In first task
context['task_instance'].xcom_push(key='report', value=report)

# In second task
report = context['task_instance'].xcom_pull(task_ids='first_task', key='report')
```

### 3. Set Appropriate Timeouts

Prevent long-running tasks:

```python
ingest = BashOperator(
    task_id='ingest',
    bash_command='...',
    execution_timeout=timedelta(hours=2),
)
```

### 4. Use Pools for Resource Management

Limit concurrent resource-intensive tasks:

```python
ingest = BashOperator(
    task_id='ingest',
    bash_command='...',
    pool='weave_ingestion',
    pool_slots=1,
)
```

### 5. Implement SLAs

Set SLAs for critical tasks:

```python
ingest = BashOperator(
    task_id='ingest',
    bash_command='...',
    sla=timedelta(hours=1),
)
```

### 6. Use TaskGroups for Organization

Group related tasks:

```python
from airflow.utils.task_group import TaskGroup

with TaskGroup('ingestion', tooltip='Document ingestion tasks') as group:
    task1 = BashOperator(...)
    task2 = BashOperator(...)
```

## Troubleshooting

### Issue: Command Not Found

**Solution**: Ensure weave-cli is in PATH:

```python
bash_command='/usr/local/bin/weave docs batch ...'
```

### Issue: Environment Variables Not Set

**Solution**: Check Airflow Variables are configured:

```bash
airflow variables set QDRANT_API_KEY "your-key"
```

### Issue: DAG Not Scheduling

**Solution**: Check start_date and schedule_interval:

```python
'start_date': datetime(2025, 1, 1),  # Must be in the past
schedule_interval='@daily',
```

## Related Documentation

- [GitHub Actions Integration](./GITHUB_ACTIONS.md)
- [Argo Workflows Integration](./ARGO_WORKFLOWS.md)
- [Apache Airflow Documentation](https://airflow.apache.org/)
- [VDB Support Matrix](../VDB_SUPPORT_MATRIX.md)
