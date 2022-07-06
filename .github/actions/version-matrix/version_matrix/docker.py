import requests
import json

from datetime import datetime


def fetch_tag_metadata(username: str, password: str, repository: str, tag: str) -> dict:
    """
    Fetch metadata for a docker repository and tag reference

    :param username:
    :param password:
    :param repository:
    :param tag:
    :return:
    """

    session = _create_session(username, password, repository)

    r = session.get(
        "https://registry-1.docker.io/v2/" + repository + "/manifests/" + tag,
        headers={"Accept": "application/vnd.docker.distribution.manifest.list.v2+json"}
    )

    tag_manifest = json.loads(r.content)
    metadata = {}

    if "manifests" in tag_manifest:
        for platform_manifest_entry in tag_manifest["manifests"]:
            digest = platform_manifest_entry["digest"]
            platform = platform_manifest_entry["platform"]
            platform_name = platform["os"] + "/" + platform["architecture"]

            r = session.get(
                "https://registry-1.docker.io/v2/" + repository + "/manifests/" + digest,
                headers={"Accept": "application/vnd.docker.distribution.manifest.list.v2+json"}
            )

            platform_manifest = json.loads(r.content)
            layer_index = len(platform_manifest["layers"]) - 1
            layer_blob = platform_manifest["layers"][layer_index]["digest"]

            r = session.head(
                "https://registry-1.docker.io/v2/" + repository + "/blobs/" + layer_blob,
                allow_redirects=True
            )

            last_modified = r.headers["Last-Modified"]
            metadata[platform_name] = datetime.strptime(last_modified, "%a, %d %b %Y %H:%M:%S %Z")

    return metadata


def _create_session(username: str, password: str, repository: str) -> requests.Session:
    """
    Create a new HTTP requests session object with the bearer token
    generated from authenticating the user from the credentials provided

    :param username:
    :param password:
    :param repository:
    :return:
    """

    r = requests.post("https://auth.docker.io/token", data={
        "grant_type": "password",
        "client_id": "dockerengine",
        "service": "registry.docker.io",
        "username": username,
        "password": password,
        "scope": "repository:" + repository + ":pull",
    })

    if r.status_code != 200:
        raise Exception("Authentication with server failed")

    auth = json.loads(r.content)
    session = requests.Session()
    session.headers = {"Authorization": "Bearer " + auth["access_token"]}

    return session
