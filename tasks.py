# To use this script you must have the following environment variables set:
#   AWS_ACCESS_KEY_ID
#   AWS_SECRET_ACCESS_KEY
# as explained in AWS documentation: https://boto3.amazonaws.com/v1/documentation/api/latest/guide/credentials.html
import json
import os
import traceback
import mimetypes

import boto3
import requests
from invoke import task
from invocations.console import confirm
from six.moves import SimpleHTTPServer, socketserver

from rds import scrape as rds_scrape
from cache import scrape as cache_scrape
from redshift import scrape as redshift_scrape
from opensearch import scrape as opensearch_scrape
from render import render
from render import build_sitemap
from render import about_page
from scrape import scrape

BUCKET_NAME = "www.ec2instances.info"

abspath = lambda filename: os.path.join(
    os.path.abspath(os.path.dirname(__file__)), filename
)

HTTP_HOST = os.getenv("HTTP_HOST", "127.0.0.1")
HTTP_PORT = os.getenv("HTTP_PORT", "8080")
REMOTE_WEBSITE_DATA_PREFIX = os.getenv("REMOTE_WEBSITE_DATA_PREFIX","https://cloud-instances.info")

def fetch_from_website_and_write_to_file(file_path):
    remote_url = f"{REMOTE_WEBSITE_DATA_PREFIX}/{file_path}"
    file_path = f"www/{file_path}"
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    try:
        response = requests.get(remote_url)
        response.raise_for_status()
        data = response.json()
        with open(file_path, "w") as f:
            f.write(json.dumps(data))
        print(f"INFO: data fetched from: {remote_url} and saved to local file: {file_path}")
        return data

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Network error while fetching data: {e}")
        return None
    except json.JSONDecodeError as e:
        print(f"ERROR: Invalid JSON received: {e}")
        return None
    except Exception as e:
        print(f"ERROR: Unexpected error: {e}")
        return None


@task
def build(c, refresh_data=False):
    """Scrape AWS sources for data and build the site"""
    scrape_ec2(c, refresh_data)
    scrape_rds(c, refresh_data)
    scrape_elasticache(c, refresh_data)
    scrape_redshift(c, refresh_data)
    scrape_opensearch(c, refresh_data)
    render_html(c)


@task
def scrape_ec2(c, refresh_data):
    """Scrape EC2 data from AWS and save to local file"""
    ec2_file = "instances.json"
    if not refresh_data:
        result = fetch_from_website_and_write_to_file(ec2_file)
        if result is None:
            print("Unable to fetch data, proceed to scrap")
        else:
        # data is fetched from website, no need to scrap aws
            return

    try:
        scrape(ec2_file)
    except Exception as e:
        print("ERROR: Unable to scrape EC2 data")
        print(traceback.print_exc())


@task
def scrape_rds(c, refresh_data):
    """Scrape RDS data from AWS and save to local file"""
    rds_file = "rds/instances.json"
    if not refresh_data:
        result = fetch_from_website_and_write_to_file(rds_file)
        if result is None:
            print("Unable to fetch data, proceed to scrap")
        else:
            # data is fetched from website, no need to scrap aws
            return

    try:
        rds_scrape(rds_file)
    except Exception as e:
        print("ERROR: Unable to scrape RDS data")
        print(traceback.print_exc())


@task
def scrape_elasticache(c, refresh_data):
    """Scrape Cache instance data from AWS and save to local file"""
    elasticache_file = "cache/instances.json"
    if not refresh_data:
        result = fetch_from_website_and_write_to_file(elasticache_file)
        if result is None:
            print("Unable to fetch data, proceed to scrap")
        else:
            # data is fetched from website, no need to scrap aws
            return

    try:
        cache_scrape(elasticache_file)
    except Exception as e:
        print("ERROR: Unable to scrape Cache data")
        print(traceback.print_exc())


@task
def scrape_redshift(c, refresh_data):
    """Scrape Redshift instance data from AWS and save to local file"""
    redshift_file = "redshift/instances.json"
    if not refresh_data:
        result = fetch_from_website_and_write_to_file(redshift_file)
        if result is None:
            print("Unable to fetch data, proceed to scrap")
        else:
            # data is fetched from website, no need to scrap aws
            return

    try:
        redshift_scrape(redshift_file)
    except Exception as e:
        print("ERROR: Unable to scrape Redshift data")
        print(traceback.print_exc())


@task
def scrape_opensearch(c, refresh_data):
    """Scrape OpenSearch instance data from AWS and save to local file"""
    opensearch_file = "opensearch/instances.json"
    if not refresh_data:
        result = fetch_from_website_and_write_to_file(opensearch_file)
        if result is None:
            print("Unable to fetch data, proceed to scrap")
        else:
            # data is fetched from website, no need to scrap aws
            return
    try:
        opensearch_scrape(opensearch_file)
    except Exception as e:
        print("ERROR: Unable to scrape OpenSearch data")
        print(traceback.print_exc())


@task
def serve(c):
    class MyHandler(SimpleHTTPServer.SimpleHTTPRequestHandler):
        def do_GET(self):
            # The URL does not include ".html". Add it to serve the file for dev
            if "/aws/" in self.path:
                if "?" in self.path:
                    self.path = (
                        self.path.split("?")[0] + ".html?" + self.path.split("?")[1]
                    )
                else:
                    self.path += ".html"
            print(self.path)
            SimpleHTTPServer.SimpleHTTPRequestHandler.do_GET(self)

    """Serve site contents locally for development"""
    os.chdir("www/")
    httpd = socketserver.TCPServer((HTTP_HOST, int(HTTP_PORT)), MyHandler)
    print(
        "Serving on http://{}:{}".format(
            httpd.socket.getsockname()[0], httpd.socket.getsockname()[1]
        )
    )
    httpd.serve_forever()


@task
def render_html(c):
    """Render HTML but do not update data from Amazon"""
    sitemap = []
    sitemap.extend(render("www/instances.json", "in/index.html.mako", "www/index.html"))
    sitemap.extend(
        render("www/rds/instances.json", "in/rds.html.mako", "www/rds/index.html")
    )
    sitemap.extend(
        render("www/cache/instances.json", "in/cache.html.mako", "www/cache/index.html")
    )
    sitemap.extend(
        render(
            "www/redshift/instances.json",
            "in/redshift.html.mako",
            "www/redshift/index.html",
        )
    )
    sitemap.extend(
        render(
            "www/opensearch/instances.json",
            "in/opensearch.html.mako",
            "www/opensearch/index.html",
        )
    )
    sitemap.append(about_page())
    build_sitemap(sitemap)


@task
def bucket_create(c):
    """Creates the S3 bucket used to host the site"""
    s3 = boto3.client("s3")

    # Create bucket with public read access
    s3.create_bucket(Bucket=BUCKET_NAME, ACL="public-read")

    # Configure website hosting
    s3.put_bucket_website(
        Bucket=BUCKET_NAME,
        WebsiteConfiguration={
            "IndexDocument": {"Suffix": "index.html"},
            "ErrorDocument": {"Key": "error.html"},
        },
    )

    print(f"Bucket {BUCKET_NAME!r} created.")


@task
def bucket_delete(c):
    """Deletes the S3 bucket used to host the site"""
    # Get bucket name from environment variable or use default
    BUCKET_NAME = os.environ.get("BUCKET_NAME", "www.ec2instances.info")

    if not confirm(f"Are you sure you want to delete the bucket {BUCKET_NAME!r}?"):
        print("Aborting at user request.")
        exit(1)

    # Determine if we're using R2 or S3 based on environment variables
    if os.environ.get("R2_ACCOUNT_ID"):
        # Using R2
        print(f"Deleting Cloudflare R2 bucket {BUCKET_NAME}")
        endpoint_url = (
            f"https://{os.environ.get('R2_ACCOUNT_ID')}.r2.cloudflarestorage.com"
        )
        s3 = boto3.client(
            "s3",
            endpoint_url=endpoint_url,
            aws_access_key_id=os.environ.get("R2_ACCESS_KEY_ID"),
            aws_secret_access_key=os.environ.get("R2_SECRET_ACCESS_KEY"),
            region_name="auto",
        )
        # R2 doesn't support ACL
        extra_args_base = {}
    else:
        # Using AWS S3
        print(f"Deleting AWS S3 bucket {BUCKET_NAME}")
        s3 = boto3.client("s3")

    bucket = s3.Bucket(BUCKET_NAME)

    # Delete all objects in the bucket first
    bucket.objects.all().delete()

    # Then delete the bucket
    bucket.delete()

    print(f"Bucket {BUCKET_NAME!r} deleted.")


@task
def deploy(c, root_dir="www", max_workers=30):
    """Deploy current content to Cloudflare R2 or S3 with parallel uploads"""
    import concurrent.futures

    # Get bucket name from environment variable or use default
    BUCKET_NAME = os.environ.get("BUCKET_NAME", "www.ec2instances.info")

    # Determine if we're using R2 or S3 based on environment variables
    if os.environ.get("R2_ACCOUNT_ID"):
        # Using R2
        print(f"Deploying to Cloudflare R2 bucket: {BUCKET_NAME}")
        endpoint_url = (
            f"https://{os.environ.get('R2_ACCOUNT_ID')}.r2.cloudflarestorage.com"
        )
        s3 = boto3.client(
            "s3",
            endpoint_url=endpoint_url,
            aws_access_key_id=os.environ.get("R2_ACCESS_KEY_ID"),
            aws_secret_access_key=os.environ.get("R2_SECRET_ACCESS_KEY"),
            region_name="auto",
        )
        # R2 doesn't support ACL
        extra_args_base = {}
    else:
        # Using AWS S3
        print(f"Deploying to AWS S3 bucket: {BUCKET_NAME}")
        s3 = boto3.client("s3")
        # S3 requires ACL for public access
        extra_args_base = {"ACL": "public-read"}

    # Collect all files to upload
    upload_tasks = []
    for root, dirs, files in os.walk(root_dir):
        for name in files:
            if name.startswith("."):
                continue

            local_path = os.path.join(root, name)
            remote_path = local_path[len(root_dir) + 1 :]
            upload_tasks.append((local_path, remote_path, name))

    total_files = len(upload_tasks)
    print(f"Uploading {total_files} files to {BUCKET_NAME}...")

    # Function to handle a single file upload
    def upload_file(task):
        local_path, remote_path, name = task
        uploads = []

        try:
            # Set content type based on file extension
            content_type = (
                "text/html"
                if name.endswith(".html")
                else mimetypes.guess_type(local_path)[0]
            )

            # Prepare extra args with content type
            extra_args = extra_args_base.copy()
            if content_type:
                extra_args["ContentType"] = content_type

            # Define the keys and labels for upload
            upload_targets = []

            # Handle index.html files specially
            if name == "index.html":
                # Upload with original path (e.g., "rds/index.html" or "index.html" for the root index file)
                upload_targets.append((remote_path, "Index HTML"))

                # For directory index files, like "rds/index.html" also create clean directory URLs
                dir_path = os.path.dirname(remote_path)
                if dir_path:
                    # For subdirectory index.html (e.g., "rds/index.html")
                    # Create "rds/" and "rds" URLs
                    upload_targets.append((dir_path + "/", "Directory URL"))
                    upload_targets.append((dir_path, "Directory URL (no slash)"))
                # For root index.html, don't create a "/" URL - that's not supported by R2. We handled that using a simple redirect Cloudflare Worker.
            # Regular HTML files (not index.html)
            elif name.endswith(".html"):
                # Get clean path without extension
                clean_path = remote_path[:-5]

                upload_targets = [
                    (remote_path, "Original URL"),  # With .html extension
                    (clean_path, "Clean URL"),  # Without .html extension
                    (clean_path + "/", "Directory URL"),  # With trailing slash
                ]
            # Non-HTML files - upload with standard path
            else:
                upload_targets = [(remote_path, "Standard")]

            # Upload all versions of the file
            for target_key, label in upload_targets:
                s3.upload_file(
                    Filename=local_path,
                    Bucket=BUCKET_NAME,
                    Key=target_key,
                    ExtraArgs=extra_args,
                )
                uploads.append((target_key, label))

            return local_path, uploads, None
        except Exception as e:
            return local_path, None, str(e)

    # Upload files in parallel
    success_count = 0
    error_count = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        future_to_task = {
            executor.submit(upload_file, task): task for task in upload_tasks
        }

        for future in concurrent.futures.as_completed(future_to_task):
            local_path, uploads, error = future.result()
            if error:
                print(f"ERROR uploading {local_path}: {error}")
                error_count += 1
            else:
                for path, type_label in uploads:
                    print(f"✓ {local_path} -> {BUCKET_NAME}/{path} ({type_label})")
                success_count += 1
    print(f"\nDeployment completed: {success_count} successful, {error_count} failed")


@task(default=True)
def update(c):
    """Build and deploy the site"""
    build(c)
    deploy(c)
