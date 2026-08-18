import requests
import os

url = os.getenv("CHAT_SERVER_APP_API_BASE_URL")

def send_request(body, **kwargs):
    args_line = ""
    if kwargs:
        pairs = []
        for key, val in kwargs.items():
            if isinstance(val, list):
                for obj in val:
                    pairs.append(f"{key}={obj}")
            else:
                pairs.append(f"{key}={val}")
        args_line = "?" + "&".join(pairs)

    request_line = f"{url}/api/v1/{body}{args_line}"
    response = requests.get(request_line)
    return response.json()

