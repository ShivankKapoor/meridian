#!/bin/bash
set -e

podman stop Meridian
podman rm Meridian
podman rmi meridian
