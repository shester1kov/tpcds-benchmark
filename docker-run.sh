#/bin/bash

docker run --rm \
  -v $(pwd)/config/config.yaml:/home/dyarn/config/config.yaml:ro \
  -v $(pwd)/tpcds_simple_queries:/home/dyarn/queries:ro \
  -v $(pwd)/results:/home/dyarn/results \
  -v $(pwd)/cacerts.pem:/home/dyarn/cacerts.pem \
  tpcds-benchmark-yarn:latest