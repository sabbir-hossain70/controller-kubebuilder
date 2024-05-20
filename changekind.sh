#!/bin/bash

OLD_KIND="customkinds"
NEW_KIND="bookservers"

# Update kind in API types
find ./api -type f -name "*.go" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
find ./cmd -type f -name "*.go" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
find ./config -type f -name "*.go" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
# Update kind in controllers
find ./internal -type f -name "*.go" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
# Update kind in CRD manifests
find ./config/crd/bases -type f -name "*.yaml" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
find ./config/crd/* -type f -name "*.yaml" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +
# Update kind in RBAC manifests
find ./config/rbac -type f -name "*.yaml" -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +


# update in all
find . -type f \( -name "*.go" -o -name "*.yaml" \) -exec sed -i "s/${OLD_KIND}/${NEW_KIND}/g" {} +

# Regenerate code and manifests
make generate
make manifests
