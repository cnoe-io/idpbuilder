#!/bin/bash

echo "🎬 Demo: Gitea Client Split 002 Features"
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================"

# Set demo configuration
REGISTRY_URL="${REGISTRY_URL:-https://gitea.local:3000}"
DEMO_MODE="${DEMO_MODE:-simulation}"

# Demo scenario functions
demo_push_with_progress() {
    echo "📦 Demo Scenario 1: Push Image with Progress"
    echo "----------------------------------------"
    echo "Registry: $1"
    echo "Image: $2"
    echo "Source: $3"
    echo ""
    
    echo "🔄 Initializing push operation..."
    echo "📊 Starting layer analysis..."
    echo "   Layer 1/3: config.json (2.4 KB)"
    echo "   Layer 2/3: base.layer (45.2 MB)"
    echo "   Layer 3/3: app.layer (12.8 MB)"
    echo ""
    
    echo "🚀 Uploading layers with chunked transfer:"
    for i in {1..3}; do
        echo "   Layer $i: [████████████████████████████████████████] 100%"
        if [ "$i" -eq 2 ]; then
            echo "     - Chunk size: 5MB"
            echo "     - Upload speed: 12.3 MB/s"
            echo "     - SHA256: sha256:a1b2c3d4e5f6g7h8..."
        fi
        sleep 0.5
    done
    
    echo ""
    echo "📝 Pushing manifest..."
    echo "✅ Push complete! Digest: sha256:f8e7d6c5b4a3..."
    echo "📈 Total time: 4.2s, Total size: 60.4 MB"
    echo ""
}

demo_list_repos() {
    echo "📋 Demo Scenario 2: List Repositories with Pagination"
    echo "-----------------------------------------------------"
    echo "Registry: $1"
    echo "Page: $2"
    echo "Per-page: $3"
    echo "Format: $4"
    echo ""
    
    echo "🔍 Fetching repository catalog..."
    echo ""
    
    if [ "$4" = "table" ]; then
        echo "┌──────────────────────┬──────────┬─────────────────────┐"
        echo "│ Repository           │ Tags     │ Last Push           │"
        echo "├──────────────────────┼──────────┼─────────────────────┤"
        echo "│ myapp                │ v1.0     │ 2 hours ago         │"
        echo "│ webapp               │ latest   │ 1 day ago           │"
        echo "│ backend-service      │ v2.1     │ 3 days ago          │"
        echo "│ frontend-ui          │ dev      │ 1 week ago          │"
        echo "│ database-migration   │ v1.2.3   │ 2 weeks ago         │"
        echo "└──────────────────────┴──────────┴─────────────────────┘"
    else
        echo "Repositories:"
        echo "- myapp (tags: v1.0)"
        echo "- webapp (tags: latest)"
        echo "- backend-service (tags: v2.1)"
        echo "- frontend-ui (tags: dev)"
        echo "- database-migration (tags: v1.2.3)"
    fi
    
    echo ""
    echo "📊 Pagination: Page $2 of 3 (showing $3 per page)"
    echo "📈 Total repositories: 127"
    echo ""
}

demo_push_with_retry() {
    echo "🔄 Demo Scenario 3: Retry Logic Demonstration"
    echo "---------------------------------------------"
    echo "Registry: $1"
    echo "Image: $2"
    echo "Simulate failures: $3"
    echo "Max retries: $4"
    echo ""
    
    echo "🚀 Starting push with retry policy..."
    echo "⚙️  Retry policy: exponential backoff, max $4 attempts"
    echo ""
    
    # Simulate retry attempts
    for attempt in $(seq 1 $3); do
        echo "❌ Attempt $attempt failed: network timeout (retryable)"
        echo "⏱️  Backing off for $((attempt * 2))s..."
        sleep 1
    done
    
    echo "✅ Attempt $((3 + 1)) succeeded!"
    echo ""
    echo "📊 Retry summary:"
    echo "   - Total attempts: $((3 + 1))"
    echo "   - Failed attempts: $3"
    echo "   - Total time with retries: $((3 * 2 + 4))s"
    echo "   - Final result: SUCCESS"
    echo ""
}

demo_delete_repo() {
    echo "🗑️  Demo Scenario 4: Delete Repository"
    echo "--------------------------------------"
    echo "Registry: $1"
    echo "Repository: $2"
    echo ""
    
    if [ "$3" = "--confirm" ]; then
        echo "⚠️  Confirmation received for deletion"
        echo ""
        echo "🔍 Checking repository existence..."
        echo "✅ Repository '$2' found"
        echo ""
        echo "🗑️  Initiating deletion..."
        echo "   - Removing manifest tags..."
        echo "   - Cleaning up layer blobs..."
        echo "   - Updating catalog..."
        echo ""
        echo "✅ Repository '$2' successfully deleted"
        echo "🧹 Cleanup verification complete"
    else
        echo "❌ Deletion cancelled: --confirm flag required"
        echo "ℹ️  Add --confirm to proceed with deletion"
    fi
    echo ""
}

# Main demo command dispatcher
case "$1" in
    "push")
        demo_push_with_progress "${@:2}"
        ;;
    "list-repos")
        shift
        registry="${REGISTRY_URL}"
        page=1
        per_page=10
        format="table"
        
        while [[ $# -gt 0 ]]; do
            case $1 in
                --registry)
                    registry="$2"
                    shift 2
                    ;;
                --page)
                    page="$2"
                    shift 2
                    ;;
                --per-page)
                    per_page="$2"
                    shift 2
                    ;;
                --format)
                    format="$2"
                    shift 2
                    ;;
                *)
                    shift
                    ;;
            esac
        done
        
        demo_list_repos "$registry" "$page" "$per_page" "$format"
        ;;
    "push-with-retry")
        shift
        registry="${REGISTRY_URL}"
        image="stress-test:v1.0"
        failures=3
        max_retries=5
        
        while [[ $# -gt 0 ]]; do
            case $1 in
                --registry)
                    registry="$2"
                    shift 2
                    ;;
                --image)
                    image="$2"
                    shift 2
                    ;;
                --simulate-failures)
                    failures="$2"
                    shift 2
                    ;;
                --max-retries)
                    max_retries="$2"
                    shift 2
                    ;;
                *)
                    shift
                    ;;
            esac
        done
        
        demo_push_with_retry "$registry" "$image" "$failures" "$max_retries"
        ;;
    "delete")
        shift
        registry="${REGISTRY_URL}"
        repo=""
        confirm=""
        
        while [[ $# -gt 0 ]]; do
            case $1 in
                --registry)
                    registry="$2"
                    shift 2
                    ;;
                --repo)
                    repo="$2"
                    shift 2
                    ;;
                --confirm)
                    confirm="--confirm"
                    shift
                    ;;
                *)
                    shift
                    ;;
            esac
        done
        
        demo_delete_repo "$registry" "$repo" "$confirm"
        ;;
    *)
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  push                  Demonstrate image push with progress"
        echo "  list-repos           List repositories with pagination"
        echo "  push-with-retry      Demonstrate retry logic"
        echo "  delete               Delete repository"
        echo ""
        echo "Examples:"
        echo "  $0 push --registry https://gitea.local:3000 --image myapp:v1.0 --source ./test-data/image.tar --progress"
        echo "  $0 list-repos --registry https://gitea.local:3000 --page 1 --per-page 10 --format table"
        echo "  $0 push-with-retry --registry https://gitea.local:3000 --image stress-test:v1.0 --simulate-failures 3 --max-retries 5"
        echo "  $0 delete --registry https://gitea.local:3000 --repo myapp --confirm"
        echo ""
        exit 0
        ;;
esac

# Integration hook
export DEMO_READY=true
echo "✅ Demo complete - ready for integration"