import logging
import argparse
import sys
import os
import json

from version_matrix import matrix_builder

parser = argparse.ArgumentParser()
parser.add_argument("repository")
parser.add_argument("--verbose", "-v", action="count", default=0)
args = parser.parse_args()

log_level = logging.WARNING - (2 if args.verbose > 2 else args.verbose) * 10
log_format = "%(asctime)s [%(levelname)s]: %(message)s [%(threadName)s]"
logging.basicConfig(format=log_format, level=log_level, datefmt="%H:%M:%S", stream=sys.stderr)

matrix = matrix_builder.build_matrix(
    username=os.environ.get("REGISTRY_USERNAME"),
    password=os.environ.get("REGISTRY_PASSWORD"),
    repository=args.repository
)

print(json.dumps(matrix))
