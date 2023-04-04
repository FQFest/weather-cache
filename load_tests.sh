!#/bin/bash

siege \
  --concurrent=100 \
  --time 5s \
  https://weather-cache-vznq5rz67q-uc.a.run.app

# Add Auth header
# --header="Authorization: bearer $(gcloud auth print-identity-token)"
