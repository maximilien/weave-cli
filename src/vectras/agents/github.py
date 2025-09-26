"""
GitHub Agent using OpenAI Agents SDK

This demonstrates how to migrate from the custom agent implementation
to using the OpenAI Agents SDK for better tool management, handoffs, and tracing.
"""

import os
from datetime import datetime
from typing import Any, Dict, List, Optional

import httpx

# OpenAI Agents SDK imports
from agents import Agent, Runner
from agents.tool import function_tool as tool
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

# Import the common response type function
from .base_agent import determine_response_type_with_llm


class GitHubIntegration:
    """GitHub API integration using the OpenAI Agents SDK tools."""

    def __init__(self, token: str, repo_owner: str, repo_name: str):
        self.token = token
        self.repo_owner = repo_owner
        self.repo_name = repo_name
        self.base_url = f"https://api.github.com/repos/{repo_owner}/{repo_name}"
        self.headers = {
            "Authorization": f"token {token}",
            "Accept": "application/vnd.github.v3+json",
        }

    async def create_branch(self, branch_name: str, base_branch: str = "main") -> bool:
        """Create a new branch from the specified base branch."""
        try:
            # Get the SHA of the base branch
            ref_url = f"{self.base_url}/git/ref/heads/{base_branch}"
            async with httpx.AsyncClient() as client:
                ref_response = await client.get(ref_url, headers=self.headers)
                ref_response.raise_for_status()
                ref_data = ref_response.json()
                sha = ref_data["object"]["sha"]

                # Create the new branch
                create_url = f"{self.base_url}/git/refs"
                create_data = {"ref": f"refs/heads/{branch_name}", "sha": sha}
                create_response = await client.post(
                    create_url, headers=self.headers, json=create_data
                )

                if create_response.status_code == 201:
                    return True
                else:
                    print(
                        f"Failed to create branch: {create_response.status_code} - {create_response.text}"
                    )
                    return False
        except Exception as e:
            print(f"Error creating branch: {str(e)}")
            return False

    async def commit_files(self, branch_name: str, files: List[str], commit_message: str) -> bool:
        """Commit files to a branch."""
        try:
            async with httpx.AsyncClient() as client:
                # Get the current commit SHA from the branch
                branch_url = f"{self.base_url}/git/refs/heads/{branch_name}"
                branch_response = await client.get(branch_url, headers=self.headers)
                branch_response.raise_for_status()
                branch_data = branch_response.json()
                current_commit_sha = branch_data["object"]["sha"]
                
                print(f"DEBUG: Current commit SHA: {current_commit_sha}")

                # Get the current tree SHA
                commit_url = f"{self.base_url}/git/commits/{current_commit_sha}"
                commit_response = await client.get(commit_url, headers=self.headers)
                commit_response.raise_for_status()
                commit_data = commit_response.json()
                current_tree_sha = commit_data["tree"]["sha"]
                
                print(f"DEBUG: Current tree SHA: {current_tree_sha}")

                # Create a new tree with the actual file changes
                new_tree_sha = await self._create_tree_with_changes(client, current_tree_sha, files)
                if not new_tree_sha:
                    print("DEBUG: Failed to create new tree with changes")
                    return False
                
                print(f"DEBUG: Created new tree SHA: {new_tree_sha}")
                
                new_commit_url = f"{self.base_url}/git/commits"
                new_commit_data = {
                    "message": commit_message,
                    "tree": new_tree_sha,
                    "parents": [current_commit_sha],
                }
                
                print(f"DEBUG: Creating commit with data: {new_commit_data}")
                new_commit_response = await client.post(
                    new_commit_url, headers=self.headers, json=new_commit_data
                )

                if new_commit_response.status_code == 201:
                    new_commit_sha = new_commit_response.json()["sha"]
                    print(f"DEBUG: Created new commit: {new_commit_sha}")
                    
                    # Update the branch reference to point to the new commit
                    update_ref_url = f"{self.base_url}/git/refs/heads/{branch_name}"
                    update_ref_data = {
                        "sha": new_commit_sha,
                        "force": False
                    }
                    update_response = await client.patch(
                        update_ref_url, headers=self.headers, json=update_ref_data
                    )
                    
                    if update_response.status_code == 200:
                        print(f"DEBUG: Successfully updated branch {branch_name}")
                        return True
                    else:
                        print(f"DEBUG: Failed to update branch: {update_response.status_code} - {update_response.text}")
                        return False
                else:
                    print(f"DEBUG: Failed to create commit: {new_commit_response.status_code} - {new_commit_response.text}")
                    return False
        except Exception as e:
            print(f"DEBUG: Error committing files: {str(e)}")
            return False

    async def _create_tree_with_changes(self, client, base_tree_sha: str, files: List[str]) -> Optional[str]:
        """Create a new tree with the actual file changes."""
        try:
            import subprocess
            import base64
            
            # Get the current project directory
            current_dir = None
            try:
                response = await client.get("http://localhost:8121/config", timeout=5.0)
                if response.status_code == 200:
                    data = response.json()
                    project_info = data.get("project_info", {})
                    if project_info and "directory" in project_info:
                        current_dir = project_info["directory"]
            except Exception as e:
                print(f"DEBUG: Error getting project directory from API: {e}")
            
            if not current_dir:
                import os
                current_dir = os.getcwd()
            
            print(f"DEBUG: Creating tree with changes in directory: {current_dir}")
            
            # Get the current tree to understand the structure
            tree_url = f"{self.base_url}/git/trees/{base_tree_sha}?recursive=1"
            tree_response = await client.get(tree_url, headers=self.headers)
            tree_response.raise_for_status()
            tree_data = tree_response.json()
            
            # Create tree entries with updated files
            tree_entries = []
            
            # Add all existing files from the base tree
            for item in tree_data.get("tree", []):
                if item["type"] == "blob":  # Only include files, not directories
                    tree_entries.append({
                        "path": item["path"],
                        "mode": item["mode"],
                        "type": "blob",
                        "sha": item["sha"]
                    })
            
            # Update or add the changed files
            for file_path in files:
                try:
                    # Read the file content
                    full_path = f"{current_dir}/{file_path}"
                    with open(full_path, 'rb') as f:
                        content = f.read()
                    
                    # Create blob with the file content
                    blob_data = {
                        "content": base64.b64encode(content).decode('utf-8'),
                        "encoding": "base64"
                    }
                    blob_response = await client.post(
                        f"{self.base_url}/git/blobs", 
                        headers=self.headers, 
                        json=blob_data
                    )
                    
                    if blob_response.status_code == 201:
                        blob_sha = blob_response.json()["sha"]
                        print(f"DEBUG: Created blob for {file_path}: {blob_sha}")
                        
                        # Update or add the file to tree entries
                        tree_entries = [entry for entry in tree_entries if entry["path"] != file_path]
                        tree_entries.append({
                            "path": file_path,
                            "mode": "100644",  # Regular file
                            "type": "blob",
                            "sha": blob_sha
                        })
                    else:
                        print(f"DEBUG: Failed to create blob for {file_path}: {blob_response.status_code}")
                        
                except Exception as e:
                    print(f"DEBUG: Error processing file {file_path}: {str(e)}")
                    continue
            
            # Create the new tree
            tree_data = {"tree": tree_entries}
            tree_response = await client.post(
                f"{self.base_url}/git/trees", 
                headers=self.headers, 
                json=tree_data
            )
            
            if tree_response.status_code == 201:
                new_tree_sha = tree_response.json()["sha"]
                print(f"DEBUG: Created new tree with {len(tree_entries)} entries: {new_tree_sha}")
                return new_tree_sha
            else:
                print(f"DEBUG: Failed to create tree: {tree_response.status_code} - {tree_response.text}")
                return None
                
        except Exception as e:
            print(f"DEBUG: Error creating tree with changes: {str(e)}")
            return None

    async def create_pull_request(
        self, branch_name: str, title: str, body: str = ""
    ) -> Optional[Dict[str, Any]]:
        """Create a pull request."""
        try:
            async with httpx.AsyncClient() as client:
                pr_url = f"{self.base_url}/pulls"
                pr_data = {"title": title, "body": body, "head": branch_name, "base": "main"}
                response = await client.post(pr_url, headers=self.headers, json=pr_data)

                if response.status_code == 201:
                    return response.json()
                else:
                    print(f"Failed to create PR: {response.status_code} - {response.text}")
                    return None
        except Exception as e:
            print(f"Error creating PR: {str(e)}")
            return None


# Global GitHub integration instance
github_integration: Optional[GitHubIntegration] = None


@tool
async def create_branch(branch_name: str, base_branch: str = "main") -> str:
    """Create a new branch from the specified base branch."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    success = await github_integration.create_branch(branch_name, base_branch)
    if success:
        return f"✅ Successfully created branch '{branch_name}' from '{base_branch}'"
    else:
        return f"❌ Failed to create branch '{branch_name}'"


@tool
async def commit_files(branch_name: str, files: List[str], commit_message: str) -> str:
    """Commit files to a branch."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    success = await github_integration.commit_files(branch_name, files, commit_message)
    if success:
        return f"✅ Successfully committed {len(files)} files to branch '{branch_name}'"
    else:
        return f"❌ Failed to commit files to branch '{branch_name}'"


@tool
async def create_pull_request(branch_name: str, title: str, body: str = "") -> str:
    """Create a pull request from a branch."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    pr_data = await github_integration.create_pull_request(branch_name, title, body)
    if pr_data:
        return f"✅ Successfully created PR #{pr_data['number']}: {pr_data['title']}\n🔗 URL: {pr_data['html_url']}"
    else:
        return f"❌ Failed to create PR from branch '{branch_name}'"


@tool
async def create_complete_pr_workflow(
    branch_name: str, files: List[str], commit_message: str, pr_title: str, pr_body: str = ""
) -> str:
    """Complete PR workflow: create branch, commit files, and create PR."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    try:
        # Step 1: Create branch
        branch_created = await github_integration.create_branch(branch_name)
        if not branch_created:
            return f"❌ Failed to create branch '{branch_name}'"

        # Step 2: Commit files
        files_committed = await github_integration.commit_files(branch_name, files, commit_message)
        if not files_committed:
            return f"❌ Failed to commit files to branch '{branch_name}'"

        # Step 3: Create PR
        pr_data = await github_integration.create_pull_request(branch_name, pr_title, pr_body)
        if pr_data:
            return f"✅ Complete PR workflow successful!\n🔗 PR #{pr_data['number']}: {pr_data['title']}\n🌐 URL: {pr_data['html_url']}"
        else:
            return f"❌ Failed to create PR from branch '{branch_name}'"

    except Exception as e:
        return f"❌ Error in PR workflow: {str(e)}"


@tool
async def list_branches() -> str:
    """List all branches in the repository."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    try:
        async with httpx.AsyncClient() as client:
            branches_url = f"{github_integration.base_url}/branches"
            response = await client.get(branches_url, headers=github_integration.headers)
            response.raise_for_status()
            branches = response.json()

            branch_list = [f"- {branch['name']}" for branch in branches]
            return f"📋 Available branches:\n{chr(10).join(branch_list)}"
    except Exception as e:
        return f"❌ Error listing branches: {str(e)}"


@tool
async def validate_files_exist(files: List[str]) -> str:
    """Validate that the specified files exist in the repository."""
    import os

    missing_files = []
    existing_files = []

    for file_path in files:
        if os.path.exists(file_path):
            existing_files.append(file_path)
        else:
            missing_files.append(file_path)

    if missing_files:
        return f"❌ Some files do not exist: {', '.join(missing_files)}\n✅ Existing files: {', '.join(existing_files)}"
    else:
        return f"✅ All files exist: {', '.join(existing_files)}"


@tool
async def get_current_changes() -> str:
    """Get current project changes and infer PR details."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    try:
        import subprocess
        import httpx
        
        # Get current project directory from API
        current_dir = None
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get("http://localhost:8121/config", timeout=5.0)
                if response.status_code == 200:
                    data = response.json()
                    project_info = data.get("project_info", {})
                    if project_info and "directory" in project_info:
                        current_dir = project_info["directory"]
                        print(f"DEBUG: Using project directory from API: {current_dir}")
        except Exception as e:
            print(f"DEBUG: Error getting project directory from API: {e}")
        
        # Fallback to current working directory
        if not current_dir:
            import os
            current_dir = os.getcwd()
            print(f"DEBUG: Using current working directory: {current_dir}")
        
        # Check git status
        print(f"DEBUG: Checking git status in directory: {current_dir}")
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            capture_output=True,
            text=True,
            cwd=current_dir,
            timeout=10
        )
        print(f"DEBUG: Git status result: {result.returncode}")
        print(f"DEBUG: Git status stdout: {result.stdout}")
        print(f"DEBUG: Git status stderr: {result.stderr}")
        
        if result.returncode == 0:
            files = result.stdout.strip().split('\n') if result.stdout.strip() else []
            if files and files[0]:  # Check if there are any changes
                changed_files = []
                for line in files:
                    if line.strip():
                        status = line[:2].strip()  # Remove leading space
                        filename = line[3:]
                        # Handle different git status formats
                        if status in ['M', 'A', 'D'] or line.startswith('??'):  # Modified, Added, Deleted, Untracked
                            changed_files.append(filename)
                
                if changed_files:
                    # Generate branch name based on changes
                    from datetime import datetime
                    timestamp = datetime.now().strftime('%Y%m%d-%H%M%S')
                    
                    # Infer change type from file names
                    if any('lint' in f.lower() or 'format' in f.lower() for f in changed_files):
                        branch_name = f"fix-linting-{timestamp}"
                        commit_msg = "fix(linting): auto-fix linting issues"
                        pr_title = "Auto-fix linting issues"
                    elif any('test' in f.lower() for f in changed_files):
                        branch_name = f"feat-testing-{timestamp}"
                        commit_msg = "feat(testing): add test improvements"
                        pr_title = "Add test improvements"
                    elif any('doc' in f.lower() or 'readme' in f.lower() for f in changed_files):
                        branch_name = f"docs-update-{timestamp}"
                        commit_msg = "docs: update documentation"
                        pr_title = "Update documentation"
                    else:
                        branch_name = f"fix-changes-{timestamp}"
                        commit_msg = "fix: resolve issues"
                        pr_title = "Fix issues"
                    
                    return f"""Current project changes detected:

Branch: {branch_name}
Changed files: {', '.join(changed_files)}
Commit message: {commit_msg}
PR title: {pr_title}

Ready to create PR with these details."""
                else:
                    return "ℹ️ No changes detected in current project."
            else:
                return "ℹ️ No changes detected in current project."
        else:
            return f"❌ Could not check git status: {result.stderr}"
            
    except Exception as e:
        return f"❌ Error checking current changes: {str(e)}"


@tool
async def get_repository_status() -> str:
    """Get the current status of the GitHub repository."""
    global github_integration
    if not github_integration:
        return "❌ GitHub integration not configured. Please set GITHUB_TOKEN environment variable."

    try:
        async with httpx.AsyncClient() as client:
            # Get repository info
            repo_url = f"{github_integration.base_url}"
            repo_response = await client.get(repo_url, headers=github_integration.headers)
            repo_response.raise_for_status()
            repo_data = repo_response.json()

            # Get recent PRs
            prs_url = f"{github_integration.base_url}/pulls?state=all&per_page=5"
            prs_response = await client.get(prs_url, headers=github_integration.headers)
            prs_response.raise_for_status()
            prs_data = prs_response.json()

            status = f"""## GitHub Repository Status

**Repository:** {repo_data["full_name"]}
**Description:** {repo_data.get("description", "No description")}
**Default Branch:** {repo_data["default_branch"]}
**Stars:** {repo_data["stargazers_count"]}
**Forks:** {repo_data["forks_count"]}
**Open Issues:** {repo_data["open_issues_count"]}

**Recent Pull Requests:**"""

            if prs_data:
                for pr in prs_data[:3]:
                    status += f"\n- #{pr['number']}: {pr['title']} ({pr['state']})"
            else:
                status += "\n- No pull requests found"

            return status
    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            return f"❌ Repository not found. Please check if the repository exists and you have access to it."
        elif e.response.status_code == 401:
            return f"❌ Authentication failed. Please check your GitHub token."
        elif e.response.status_code == 403:
            return f"❌ Access forbidden. Please check your GitHub token permissions."
        else:
            return f"❌ HTTP error {e.response.status_code}: {e.response.text}"
    except httpx.RequestError as e:
        return f"❌ Network error: {str(e)}"
    except Exception as e:
        return f"❌ Error getting repository status: {str(e)}"


def get_current_repo_info():
    """Get current repository info from git remote."""
    try:
        import subprocess
        result = subprocess.run(
            ["git", "remote", "get-url", "origin"],
            capture_output=True,
            text=True,
            timeout=5
        )
        if result.returncode == 0:
            remote_url = result.stdout.strip()
            print(f"DEBUG: Git remote URL: {remote_url}")
            # Extract owner/repo from git URL
            if "github.com" in remote_url:
                parts = remote_url.replace(".git", "").split("/")
                if len(parts) >= 2:
                    owner = parts[-2]
                    repo = parts[-1]
                    print(f"DEBUG: Detected repository: {owner}/{repo}")
                    return owner, repo
    except Exception as e:
        print(f"DEBUG: Error getting git remote: {e}")
    return None, None


def initialize_github_integration():
    """Initialize GitHub integration with environment variables."""
    global github_integration

    token = os.getenv("GITHUB_TOKEN")
    print(f"DEBUG: GITHUB_TOKEN = {'SET' if token else 'NOT SET'}")
    
    # Try to get repo info from environment variables first
    repo_owner = os.getenv("GITHUB_ORG")
    repo_name = os.getenv("GITHUB_REPO")
    print(f"DEBUG: Environment variables - GITHUB_ORG: {repo_owner}, GITHUB_REPO: {repo_name}")
    
    # If not set, try to detect from git remote
    if not repo_owner or not repo_name:
        print("DEBUG: Environment variables not set, detecting from git remote...")
        detected_owner, detected_repo = get_current_repo_info()
        if detected_owner and detected_repo:
            repo_owner = detected_owner
            repo_name = detected_repo
            print(f"DEBUG: Using detected repository: {repo_owner}/{repo_name}")
        else:
            # Fallback to defaults
            repo_owner = "maximilien"
            repo_name = "vectras-ai"
            print(f"DEBUG: Using fallback repository: {repo_owner}/{repo_name}")

    if token:
        github_integration = GitHubIntegration(token, repo_owner, repo_name)
        print(f"🔧 Initialized GitHub integration for {repo_owner}/{repo_name}")
    else:
        print("⚠️ No GitHub token found. GitHub integration disabled.")
        github_integration = None


# Initialize GitHub integration
initialize_github_integration()


# Create the GitHub agent using OpenAI Agents SDK
github_agent = Agent(
    name="GitHub Agent",
    instructions="""You are the Vectras GitHub Agent. You help with GitHub operations like creating branches, committing code, and creating pull requests.

Your capabilities include:
- Creating branches from existing branches
- Committing files to branches
- Creating pull requests
- Listing repository branches
- Getting repository status

When users ask for status, provide a comprehensive overview of the repository.

**CRITICAL: Handle ALL PR creation requests by using the create_complete_pr_workflow tool. Do NOT provide advice or ask for confirmation.**

**For PR Creation Requests:**
When you receive any PR creation request:
1. ALWAYS use get_current_changes tool first to detect current project state
2. Extract branch name, changed files, and summary from the query OR use detected information
3. Use the create_complete_pr_workflow tool with the detected information
4. Use appropriate commit message format based on the changes
5. Use descriptive PR title and body
6. NEVER ask for details - always infer from current project state using get_current_changes

**For Auto-Fix Changes:**
When receiving auto-fix changes from linting agent:
- Use commit message: "fix(linting): auto-fix linting issues"
- Use PR title: "Auto-fix linting issues"
- Use PR body with the provided summary

**For General Changes:**
When receiving other change requests:
- Use appropriate commit message based on the type of changes
- Use descriptive PR title based on the changes
- Use detailed PR body explaining what was changed

**For Commit Messages:**
Use descriptive, conventional commit format:
- Format: "type(scope): description"
- Examples: "fix(linting): auto-fix linting issues", "feat(api): add new endpoint", "docs(readme): update installation instructions"

**For PR Titles:**
Use clear, descriptive titles:
- Examples: "Auto-fix linting issues", "Add new API endpoint", "Update documentation"

**For PR Bodies:**
Include:
- Summary of changes
- What was fixed/added/updated
- Testing performed (if applicable)
- Any breaking changes (if any)

**CRITICAL: When asked to create a PR, you MUST take action by calling the create_complete_pr_workflow tool. Do NOT just provide advice or troubleshooting steps. Always attempt the actual PR creation workflow.**

If the workflow fails, report the error but DO NOT provide troubleshooting advice. Just state what happened and what the error was.

Always be helpful and provide clear, actionable responses.

You can use the following tools to perform GitHub operations:
- get_current_changes: Check current project state and infer PR details
- validate_files_exist: Check if files exist before committing
- create_branch: Create a new branch from a base branch
- commit_files: Commit files to a branch with a message
- create_pull_request: Create a pull request from a branch
- create_complete_pr_workflow: Complete PR workflow (branch + commit + PR)
- list_branches: List all branches in the repository
- get_repository_status: Get comprehensive repository status

If a user asks about something outside your capabilities (like code analysis, testing, or linting), you can suggest they ask the appropriate agent:
- For code analysis and fixes: Ask the Coding Agent
- For testing: Ask the Testing Agent
- For code quality and formatting: Ask the Linting Agent
- For log monitoring: Ask the Logging Monitor Agent
- For project coordination: Ask the Supervisor Agent

Format your responses in markdown for better readability.""",
    tools=[
        create_branch,
        commit_files,
        create_pull_request,
        create_complete_pr_workflow,
        validate_files_exist,
        list_branches,
        get_repository_status,
        get_current_changes,
    ],
)


# FastAPI app for web interface compatibility
app = FastAPI(
    title="Vectras GitHub Agent",
    description="GitHub operations agent",
    version="0.2.0",
)

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


class QueryRequest(BaseModel):
    query: str
    context: Optional[Dict[str, Any]] = None


class QueryResponse(BaseModel):
    status: str
    response: str
    agent_id: str = "github"
    timestamp: datetime
    metadata: Dict[str, Any]


@app.post("/query", response_model=QueryResponse)
async def query_endpoint(request: QueryRequest) -> QueryResponse:
    """Main query endpoint that uses the OpenAI Agents SDK."""
    try:
        print(f"DEBUG: GitHub agent received query: {request.query[:100]}...")

        # Run the agent using the SDK
        result = await Runner.run(github_agent, request.query)

        # Determine response type for frontend rendering using LLM when needed
        response_type = await determine_response_type_with_llm(
            "github", request.query, result.final_output
        )

        return QueryResponse(
            status="success",
            response=result.final_output,
            timestamp=datetime.now(),
            metadata={
                "model": "gpt-4o-mini",
                "capabilities": ["Branch Management", "PR Creation", "Repository Operations"],
                "response_type": response_type,
                "sdk_version": "openai-agents",
            },
        )

    except Exception as e:
        print(f"Error in GitHub agent: {str(e)}")
        return QueryResponse(
            status="error",
            response=f"Error processing query: {str(e)}",
            timestamp=datetime.now(),
            metadata={"error": str(e)},
        )


@app.get("/health")
async def health():
    return {"status": "ok", "service": "github-agent"}


@app.get("/status")
async def status():
    return {
        "agent": "GitHub Agent",
        "status": "active",
        "github_configured": github_integration is not None,
        "sdk_version": "openai-agents",
        "tools": [
            "create_branch",
            "commit_files",
            "create_pull_request",
            "create_complete_pr_workflow",
            "list_branches",
            "get_repository_status",
        ],
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8128)
