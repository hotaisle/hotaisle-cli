# From a GPU VM to an inference endpoint

This walkthrough connects the CLI's provisioning and connection commands to a vLLM smoke test on AMD hardware. It uses a small public model to check the deployment path. The CLI commands are covered by mocked API tests; the Docker/model steps require validation on an actual GPU VM and are not a published benchmark or a verified DeepSeek deployment.

## Provision a GPU VM

```bash
hotaisle setup
hotaisle doctor
hotaisle --json vm available
```

Register a public SSH key during setup. Inspect available GPU models, counts, hourly prices, and minimum reservation time. Substitute the model returned by availability rather than assuming a particular machine is in stock:

```bash
hotaisle --json vm provision --gpu-model YOUR_GPU_MODEL --gpu-count 1
```

Save the returned `.name` as `YOUR_VM`. If the request times out, inspect `hotaisle --json vm list` before retrying: provisioning continues after an HTTP timeout, particularly with `--user-data-url`.

```bash
hotaisle --json vm wait --vm YOUR_VM --timeout 5m
hotaisle vm ssh --vm YOUR_VM
```

Run the printed SSH command. `vm wait` verifies power state only; it does not establish that the guest has finished booting. If SSH is unavailable, confirm the public key is registered and retry with a deadline. Once connected, wait for guest initialization:

```bash
sudo cloud-init status --wait
```

If cloud-init fails, inspect `sudo tail -n 100 /var/log/cloud-init-output.log` before starting the model server. Check Docker is installed (`docker --version`) and the AMD GPU devices are available (`ls -l /dev/kfd /dev/dri`). Driver and runtime installation depend on the VM image; use the [vLLM AMD installation instructions](https://docs.vllm.ai/en/latest/getting_started/installation/gpu/) when those prerequisites are missing.

## Start a small model on the VM

The command below follows [vLLM's official ROCm Docker guidance](https://docs.vllm.ai/en/latest/deployment/docker/) and its [documented ROCm device configuration](https://docs.vllm.ai/en/latest/getting_started/installation/gpu/). Select a released image tag or digest compatible with the installed host driver. The current upstream tag examples can change, so pin the image after validating your chosen combination.

```bash
# Run on the GPU VM, replacing this with your validated release tag or digest.
VLLM_IMAGE='vllm/vllm-openai-rocm:YOUR_VALIDATED_TAG'
sudo docker run -d --name hotaisle-inference \
  --group-add=video \
  --cap-add=SYS_PTRACE \
  --security-opt seccomp=unconfined \
  --device /dev/kfd --device /dev/dri \
  --ipc=host \
  -v /var/cache/hotaisle-models:/root/.cache/huggingface \
  -p 127.0.0.1:8000:8000 \
  "$VLLM_IMAGE" --model Qwen/Qwen3-0.6B --host 0.0.0.0 --port 8000

sudo docker logs --tail 100 hotaisle-inference
```

Port 8000 is bound to loopback on the VM. Access it through SSH forwarding instead of exposing an unauthenticated endpoint publicly. On your local machine, get the connection fields once:

```bash
hotaisle --json vm ssh --vm YOUR_VM
```

Use the returned host and port in a separate terminal:

```bash
ssh -p YOUR_SSH_PORT -L 8000:127.0.0.1:8000 hotaisle@YOUR_SSH_HOST
```

## Verify the application

Model download and initialization can take longer than VM boot. From your local machine, wait for application health with a finite retry budget:

```bash
curl --fail --silent --show-error \
  --connect-timeout 5 --max-time 10 \
  --retry 30 --retry-delay 5 --retry-all-errors --retry-max-time 180 \
  http://127.0.0.1:8000/health

curl --fail --silent --show-error --connect-timeout 5 --max-time 10 \
  http://127.0.0.1:8000/v1/models

curl --fail --silent --show-error --connect-timeout 5 --max-time 60 \
  http://127.0.0.1:8000/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"Qwen/Qwen3-0.6B","messages":[{"role":"user","content":"Say hello in one sentence."}],"max_tokens":128}'
```

If health never succeeds, inspect `docker logs` for driver incompatibility, insufficient memory, model download failures, or unsupported architecture. To deploy another model, verify its exact repository identifier, supported runtime release, memory requirements, and parallelism settings. A name such as “DeepSeek 4.1 Flash” alone does not establish those requirements; avoid assuming it fits the single-GPU smoke-test configuration.

After validating the workflow, a maintained cloud-init file can automate guest setup with `vm provision --user-data-url HTTPS_URL`. The URL must be reachable by the provisioning service; do not publish API keys or other secrets in it. Custom user-data triggers a VM rebuild and can extend provisioning beyond the HTTP request deadline.

## Clean up

Preserve any data you need before deleting the VM:

```bash
hotaisle --json vm delete --vm YOUR_VM
hotaisle --json vm list
```

Deleting the VM ends billing, subject to minimum reservation charges. Stopping the container or running `vm stop` does not end VM billing.
