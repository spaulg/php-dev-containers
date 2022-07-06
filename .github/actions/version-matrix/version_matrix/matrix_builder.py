import datetime
import threading
import logging

from . import release_versions
from . import docker
from . import constant


def build_matrix(username: str, password: str, repository: str) -> dict:
    """
    Build the version matrix by spawning checks for validity
    of the version detected against previous builds, their age
    and if newer patch versions are available

    :param username:
    :param password:
    :param repository:
    :return: None
    """

    threads = []
    matrix = {"version": [], "include": []}
    failures = []

    for version in release_versions.list_all_versions():
        logging.info("Spawning check for version %s", version)

        thread = threading.Thread(
            target=_check_version,
            args=(version, matrix, failures, username, password, repository),
            name="BuildThread-" + version
        )

        threads.append(thread)
        thread.start()

    logging.info("Waiting for all threads to finish...")
    for thread in threads:
        thread.join()

    # Sort versions
    matrix["version"].sort()

    return matrix


def _check_version(version_number: str, matrix: dict, failures: list, username: str, password: str, repository: str):
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

    minimum_age = datetime.datetime.now() - datetime.timedelta(days=constant.MAX_AGE_IN_DAYS)
    epoch = str(int(datetime.datetime.now().timestamp()))

    try:
        version_metadata = release_versions.fetch_version_metadata(version_number)
        logging.debug(version_metadata)

        tag_metadata = docker.fetch_tag_metadata(username, password, repository, version_number)
        logging.debug(tag_metadata)

        if "platform" in tag_metadata:
            for platform in tag_metadata:
                last_modified = tag_metadata[platform]

                if last_modified < version_metadata["release_date"] or last_modified < minimum_age:
                    logging.info("Appending %s to build list", version_number)

                    _append_version(
                        version_number,
                        version_metadata,
                        matrix,
                        epoch,
                    )

                    return
        else:
            # Version missing in docker, add for first build
            _append_version(
                version_number,
                version_metadata,
                matrix,
                epoch,
            )

    except Exception:
        failures.append(version_number)


def _append_version(version_number: str, version_metadata: dict, matrix: dict, epoch: str):
    """
    Append both the nts and zts versions of a particular version to the version matrix

    :param version_number:
    :param version_metadata:
    :param matrix:
    :param epoch:
    :return:
    """

    _append_version_entry(
        version_number,
        version_metadata,
        matrix,
        epoch,
    )

    _append_version_entry(
        version_number,
        version_metadata,
        matrix,
        epoch,
        "zts",
    )


def _append_version_entry(version_number: str, version_metadata: dict, matrix: dict, epoch: str, suffix: str = None):
    """
    Append the desired version to the version matrix with the optional suffix

    :param version_number:
    :param version_metadata:
    :param matrix:
    :param epoch:
    :param suffix:
    :return:
    """

    hyphenated_suffix = ""
    if suffix is not None:
        hyphenated_suffix = "-" + suffix
    else:
        suffix = ""

    matrix["version"].append(version_number + hyphenated_suffix)
    matrix["include"].append({
        "version": version_number + hyphenated_suffix,
        "suffix": suffix,
        "package_name": "php" + version_metadata["short_version"] + suffix,
        "package_version": version_metadata["full_version"] + suffix,
        "package_upstream_filename": "php" + version_metadata["short_version"] + suffix + "_" +
                                     version_metadata["full_version"] + ".orig.tar.gz",
        "dsc_filename": "php" + version_metadata["short_version"] + suffix + "_" +
                        version_metadata["full_version"] + "-" + epoch + ".dsc",
        "full_package_version": version_metadata["full_version"] + "-" + epoch,
        "short_version": version_metadata["short_version"],
        "full_version": version_metadata["full_version"],
        "asset_url": version_metadata["asset_url"],
        "asset_filename": version_metadata["asset_filename"],
    })
