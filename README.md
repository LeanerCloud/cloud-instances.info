# EC2Instances.info

[![uses aws](https://img.shields.io/badge/uses-AWS-yellow)](https://aws.amazon.com/)
[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/release/python-380/)
[![python style: black](https://img.shields.io/badge/python%20style-black-000000.svg?style=flat-square)](https://github.com/psf/black)

![Vantage Picture](https://uploads-ssl.webflow.com/5f9ba05ba40d6414f341df34/5f9bb1764b6670c6f7739564_moutain-scene.svg)

> I was sick of comparing EC2 instance metrics and pricing on Amazon's site so I
> made EC2Instances.info.

This website was originally known as EC2Instances.info and was created by [Garret
Heaton](https://github.com/powdahound) in 2011.

For many years it has been a great resource for AWS users interested in comparing EC2 instance types by specs and prices.

The website also offers a very useful database of pricing and specs information for AWS EC2 instances which is also used by many other projects.

As of May 2025 is was forked and now hosted at [cloud-instances.info](https://cloud-instances.info), supported by
[Cristian Magherusan-Stanciu](https://LeanerCloud.com/) and developed in the open with the community of contributors.

## Project status

Vantage acquired the ec2instances.info website from the original author in 2021 and mainly used the website as part of their marketing efforts.

They actively developed it for a few years, improving it in many ways, redesigning the UI and exposing new data sources.

Unfortunately the project hasn't seen much development for more than a year, and last contributions from Vantage employees were in October 2024.

All changes over the last 6 months were contributed by the community, mainly by [Cristian Magherusan-Stanciu](https://LeanerCloud.com/), a cloud optimization specialist who uses the website daily on his optimization gigs, and since 2016 has been building multiple projects that rely on on the instance specs database.

Cristian was a major contributor to the project and co-maintainer before Vantage acquired it and occasionally kept contributing code and reporting issues also after the acquisition.

In early 2025 Vantage finally hired someone to work again on the website, but they are working behind closed doors and decided to rewrite the website from scratch, instead of addressing the ever accumulating issues.

It is not clear how the rewritten website will look like, whether it will support the same functionality and instance specs database the users are relying on.

As any rewrite, chances are it will take time for it to mature, introducing potentially new bugs, while the community keeps reporting issue and sending pull requests that Vantage isn't very responsive in addressing.

Also, some recent features such as Azure support code haven't been open sourced by Vantage, the UI is broken and nobody can fix it.

Because of all these reasons, in May 2025 Cristian decided to fork the project and create a new version of the website, which aims to be vendor neutral and developed in the open together with the rest of the community.

The plan is to rewrite all the closed source Vantage features, remove all references to Vantage that may infringe on their copyrights or trademarks (logos, icons, etc.), and to extend it further to cover additional cloud providers.

## Contributing

Improvements in the form of pull requests or ideas via issues are very welcome from anyone!

We also have a [Slack Community](https://join.slack.com/t/leanercloud/shared_invite/zt-xodcoi9j-1IcxNozXx1OW0gh_N08sjg) for anyone to join, with a dedicated support and development channel named #cloud-instances-info.

## Running locally

First, you'll need to provide credentials so that boto can access the AWS API. [See a terraform example here](./docs/terraform/iam.tf)!
Options for setting this up are described in the [boto
docs](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html).

Ensure that your IAM user has at least the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstanceTypes",
        "ec2:DescribeRegions",
        "pricing:*",
        "elasticache:DescribeEngineDefaultParameters"
      ],
      "Resource": "*"
    }
  ]
}
```

## Running in Docker (recommended)

1. Clone the repository, if not already done:

```bash
git clone https://github.com/LeanerCloud/cloud-instances.info
cd cloud-instances.info
```

2. Build a `docker` image:

```bash
docker build -t cloud-instances.info .
```

3. Run a container from the built `docker` image:

````bash
docker run -d --name some-container -p 8080:8080 cloud-instances.info

4. Open [localhost:8080](http://localhost:8080) in your browser to see it in action.

## Docker Compose

Here's how you can build and run docker image using Docker Compose (tested with Docker Compose v2):

```bash
docker-compose up
````

4. Open [localhost:8080](http://localhost:8080) in your browser to see it in action.

## Detailed local build instructions

Note: These instructions are only kept here for reference, the Docker
instructions mentioned above hide all these details and are recommended for local execution.

Make sure you have LibXML and Python development files. On Ubuntu, run `sudo apt-get install python-dev libxml2-dev libxslt1-dev libssl-dev`.

Then:

```bash
git clone https://github.com/LeanerCloud/cloud-instances.info
cd cloud-instances.info/
python3 -m venv env
source env/bin/activate
pip install -r requirements.txt
invoke build
invoke serve
open http://localhost:8080
deactivate # to exit virtualenv
```

## Requirements

- Python with virtualenv
- [Invoke](http://www.pyinvoke.org/)
- [Boto](http://boto.readthedocs.org/en/latest/)
- [Mako](http://www.makotemplates.org/)
- [lxml](http://lxml.de/)

## Build Options

Build command accepts a flag --refresh-data which will performs a full data scrap of latest pricing data from AWS

Usage

```bash
invoke build --refresh-data
```

A normal build i.e. `invoke build` fetches the latest published pricing data from cloud-instances.info for the local website build.

## Tips for Developing Locally

```
docker build --no-cache --build-arg AWS_ACCESS_KEY_ID= --build-arg AWS_SECRET_ACCESS_KEY= -t ec2instances.info .

docker run -it --rm --name ec2instances -v $(pwd):/opt/app --env HTTP_HOST=0.0.0.0 -p 8080:8080 ec2instances.info

docker exec -it ec2instances /bin/bash

# INSIDE CONTAINER
python3 render.py
sass --watch in/style.scss:www/style.css
```

## API Access

The data source is available via a free API offered by Vantage.

- To get started, create a [free API key](https://vantage.readme.io/reference/authentication).
- Review the `providers`, `services`, and `products` endpoints in the [API documentation](https://vantage.readme.io/reference/getproducts).

## Keep up-to-date

Feel free to watch/star this repo as we're looking to update the site regularly. Vantage also works on the following relevant projects:

- [The Cloud Cost Handbook](https://github.com/vantage-sh/handbook) - An
  open-source set of guides for best practices of managing cloud costs.
- [The AWS Cost Leaderboard](https://leaderboard.vantage.sh/) - A hosted site of
  the top AWS cost centers.
- [Vantage](https://vantage.sh/) - A cloud cost transparency platform.
