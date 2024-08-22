import threading
import logging

from . import release_versions
from . import constant


def build_matrix() -> dict:
    """
    Build the version matrix

    :return: None
    """

    threads = []
    matrix = {"short_sem_version": [], "include": []}
    failures = []

    for version in release_versions.list_all_versions():
        logging.info("Spawning check for version %s", version)

        thread = threading.Thread(
            target=_check_version,
            args=(version, matrix, failures),
            name="BuildThread-" + version
        )

        threads.append(thread)
        thread.start()

    logging.info("Waiting for all threads to finish...")
    for thread in threads:
        thread.join()

    # Sort versions
    matrix["short_sem_version"].sort()

    return matrix


def _check_version(version_number: str, matrix: dict, failures: list):
    """
    Check a particular version for freshness

    :param version_number:
    :param matrix:
    :param failures:
    :param :username:
    :param password:
    :param repository:
    :return:
    """

    try:
        version_metadata = release_versions.fetch_version_metadata(version_number)
        logging.debug(version_metadata)

        _append_version_entry(
            version_metadata,
            matrix,
        )

        _append_version_entry(
            version_metadata,
            matrix,
            "zts",
        )

    except Exception:
        failures.append(version_number)


def _append_version_entry(version_metadata: dict, matrix: dict, suffix: str = None):
    """
    Append the desired version to the version matrix with the optional suffix

    :param version_metadata:
    :param matrix:
    :param epoch:
    :param suffix:
    :return:
    """

    metadata = ""
    if suffix is not None:
        metadata = "+" + suffix

    matrix["short_sem_version"].append(version_metadata["short_version"] + metadata)
    matrix["include"].append({
        "short_sem_version": version_metadata["short_version"] + metadata,
        "full_sem_version": version_metadata["full_version"] + metadata,
        "short_version": version_metadata["short_version"],
    })
