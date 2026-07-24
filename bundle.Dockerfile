FROM scratch

# Copy all manifests and metadata into the bundle container
COPY bundle/manifests /manifests/
COPY bundle/metadata /metadata/

# Annotations required for OLM bundle indexers
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=olm-update-sentinel
LABEL operators.operatorframework.io.bundle.channels.v1=alpha,stable
LABEL operators.operatorframework.io.bundle.channel.default.v1=stable