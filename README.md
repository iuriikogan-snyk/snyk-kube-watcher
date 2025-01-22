# Snyk Kube-watcher

## High-Level Architecture

Kubeconfigs + Contexts
The user provides paths to one or more kubeconfig files and possibly multiple contexts. If multiple contexts and kubeconfigs are given, the application pairs each context with each kubeconfig (or a one-to-one mapping, as desired).

Build Kubernetes Clients
For each (kubeconfig, context) pair, the application uses clientcmd.BuildConfigFromFlags (or its variants) to build a rest.Config, then creates a clientset for that cluster.

Spawn Watchers
Each (kubeconfig, context) pair initializes a Shared Informer Factory (e.g., informers.NewSharedInformerFactory), focusing on Pod resources. This watcher monitors all namespaces for new or updated Pods.

Shared Informer: Pod Events

Add events occur when a new Pod is scheduled.

Update events occur when a Pod changes (e.g., container restarts, updates to labels, etc.).

Extract Container Images
For each Pod, the application iterates over pod.Spec.Containers and extracts the Image field (e.g., nginx:latest, alpine:3.17, etc.).

Global Task Queue (taskCh)
All watchers push images into a single, centralized queue so that the same concurrency and rate-limiting logic applies across all clusters.

Worker Pool
A fixed number of workers (e.g., concurrency = 5) read from the queue in parallel. Each worker ensures:

Rate Limit: Before each Snyk API call, the worker waits for a token from the rate limiter (e.g., 2 requests/second).

Retry with Backoff: On failure (HTTP error, etc.), the worker retries up to maxRetries times, doubling the wait (backoffFactor) each time.

Snyk API

Each worker issues a POST to the Snyk API endpoint (e.g., /v1/org/<ORG_ID>/container-monitoring) with the container image.

If the response is successful, Snyk now monitors that image for vulnerabilities.

Graceful Shutdown

On receiving a SIGINT (Ctrl+C) or SIGTERM, the main context is canceled.

Each watcher (informer) is signaled to stop retrieving further Pod updates.

The main routine closes the global queue once watchers exit.

Workers finish any in-flight tasks and exit, ensuring no partial scans remain.

## Implementation Outline

Parse Flags

Read -kubeconfig (repeatable) to get multiple paths.

Read -context (repeatable) to get multiple context names.

Read -org for the Snyk organization ID.

(Optionally) if no kubeconfig is provided, default to ~/.kube/config.

Create Signal-Aware Context

Use signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) to handle container orchestrator signals.

Start Worker Pool

Create a task channel (taskCh) and start concurrencygoroutines.

Each worker processes images with a rate limiter and retry logic.

Spawn Watchers

For each combination of kubeconfig + context, build a rest.Config and create a Kubernetes client.

Use a Shared Informer Factory to watch Pod resources.

On Pod add/update, extract container images and push them into the global taskCh.

Wait for Watchers

Each watcher runs until the context is canceled.

Once watchers exit, close the taskCh.

Wait for Workers

Workers exit naturally after the channel is closed and they process any remaining tasks.

The application then logs a final status and exits.
