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

    release_asset = _extract_release_url(version_number, version_metadata)

    return {
        "short_version": version_number,
        "full_version": version_metadata['version'],
        "asset_url": release_asset[0],
        "asset_filename": release_asset[1],
        "release_date": _extract_release_datetime(version_metadata),
    }


def _extract_release_url(version_number: str, version_metadata: dict) -> list:
    """
    Extract a fully qualified URL of the release source package
    and return

    :param version_number:
    :param version_metadata:
    :return:
    """

    major, minor = version_number.split(".")
    museum = version_metadata["museum"] if "museum" in version_metadata else False

    for file_metadata in version_metadata["source"]:
        if file_metadata["filename"].endswith(".tar.gz"):
            return [_expand_release_url(major, file_metadata["filename"], museum), file_metadata["filename"]]

    raise Exception("Failed to expand source file in to a fully qualified URL")


def _expand_release_url(major_version_number: str, filename: str, museum: bool) -> str:
    """
    Expand a release filename in to a fully qualified URL, based on the
    museum status

    :param major_version_number:
    :param filename:
    :param museum:
    :return:
    """

    if museum:
        return "https://museum.php.net/php{}/{}".format(major_version_number, filename)
    else:
        return "https://www.php.net/distributions/{}".format(filename)


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
