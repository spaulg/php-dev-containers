import requests
import json
from datetime import datetime

from . import constant


def list_all_versions() -> list:
    """
    Fetch the latest versions of each PHP major/minor build
    combination

    :return:
    """

    r = requests.get(constant.PHP_RELEASE_API)
    latest_releases = json.loads(r.content)

    if r.status_code != 200:
        raise Exception("Failed to fetch all versions for listing for matrix")

    for release in latest_releases:
        major_version, minor_version, patch_version = latest_releases[release]["version"].split(".")
        major_version_int = int(major_version)
        minor_version_int = int(minor_version)

        while major_version_int == constant.PHP_MIN_MAJOR_VERSION and minor_version_int >= constant.PHP_MIN_MINOR_VERSION or \
            major_version_int > constant.PHP_MIN_MAJOR_VERSION and minor_version_int >= 0:
            major_minor_version = major_version + "." + str(minor_version_int)
            minor_version_int = minor_version_int - 1
            yield major_minor_version


def fetch_version_metadata(version_number: str) -> dict:
    """
    Fetch version metadata for the version number passed

    :param version_number:
    :return:
    """

    r = requests.get(constant.PHP_RELEASE_API, params={"version": version_number})
    version_metadata = json.loads(r.content)

    if r.status_code != 200:
        raise Exception("Failed to fetch metadata for version")

    return {
        "short_version": version_number,
        "full_version": version_metadata['version'],
        "release_date": _extract_release_datetime(version_metadata),
    }


def _extract_release_datetime(version_metadata: dict) -> datetime:
    """
    Extract the release timestamp from the version metadata

    :param version_metadata:
    :return:
    """

    for date_format in ["%d %b %Y", "%d %B %Y"]:
        try:
            return datetime.strptime(version_metadata["date"], date_format)
        except ValueError:
            pass

    raise Exception("Failed to parse the datetime used in the release metadata")
