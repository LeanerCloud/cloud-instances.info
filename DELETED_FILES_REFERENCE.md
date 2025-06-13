# Files Deleted in Vantage Branch

This document contains a reference of all files that exist in our branch but were deleted in the Vantage branch. This is for reference during the rebase process.

## Summary

Total files deleted: 91 files

These files fall into several categories:
1. GitHub Actions workflows (.github/workflows/)
2. Docker configuration files
3. Detail pages Python scripts
4. Template files (in/*.mako)
5. Infrastructure/Terraform files
6. Frontend assets (Bootstrap, CSS, JS)
7. Render system (render.py)

## File List

See deleted_files.txt for the complete list.

## Important Files to Consider

### 1. Sponsors Page Template
- `in/sponsors.html.mako` - The sponsors page template that needs to be preserved

### 2. Render System
- `render.py` - The rendering system for generating HTML pages

### 3. Detail Pages
- `detail_pages_*.py` - Scripts for generating detail pages for each service

### 4. Infrastructure
- All files under `infra/` - Terraform configurations for deployment

### 5. Frontend Assets
- Bootstrap files
- Custom CSS and JS files

## Next Steps

During the rebase, we will:
1. Skip commits that delete these files
2. Preserve the sponsors functionality
3. Adapt any necessary changes to work with the Vantage codebase structure

## Deleted File Contents

### 1. in/sponsors.html.mako

This is the Mako template for the sponsors page. Key features:
- Inherits from base.mako
- Has multiple sponsor tiers: Gold, Silver, Bronze
- Includes event sponsors section
- Call-to-action sections for becoming a sponsor

```mako
<%!
  active_ = ""
%>
<%inherit file="base.mako" />

    <%block name="meta">
        <title>Sponsors - Cloud-Instances.info</title>
        <meta name="description" content="Supporting the development and maintenance of Cloud-Instances.info - Thank you to our sponsors and supporters!">
    </%block>

<%!
  active_ = ""
%>
<%inherit file="base.mako" />

    <!-- Override any height restrictions that might be cutting off content -->
    <style>
      html, body {
        height: auto !important;
        overflow: visible !important;
      }
      .main {
        height: auto !important;
        overflow: visible !important;
      }
    </style>

    <div class="mx-2 my-4" style="overflow: visible;">
      <div class="row">
        <div class="col-12">
          <div class="row g-4 justify-content-center">
            <div class="col-lg-8">
              <div class="text-center mb-5">
                   <p class="lead text-muted">
                <h1 class="display-4 fw-bold mb-3">Why Sponsor Us?</h1>

                The open source software powering Cloud-Instances.info has been built for many years by a few dedicated volunteers working on it in their limited spare time. The main author eventually gave up and sold it to Vantage, becoming a very important marketing asset.
                <br>
                <br>
                    After a few years of active development by Vantage, it became neglected and development happened only in a private branch for a while.
                <br>
                <br>
                    @cristim, a former co-maintainer before the Vantage acquisition decided to fork the open source code and now trying to bring it back into a community driven project, actively maintained and developed in the open, with the main page free of marketing messaging since most users use it for work.
                <br>
                <br>
                    But at the same time we want to fund a few people to work on it consistently, and for that we accept sponsorships of all kinds.
                <br>
                <br>
                    We really appreciate any help to keep this tool actively maintained and vendor-neutral and are very thankful to our current sponsors!
                </p>

              </div>
            </div>
          </div>

          <!-- Gold Tier -->
          <div class="mb-5">
            <h2 class="h3 fw-bold mb-4 text-center">
              <span class="badge text-dark me-2" style="background-color: #ffd700;">Gold Sponsors</span>
            </h2>

            <div class="row justify-content-center">
              <div class="col-lg-10">
                <div class="row g-4 justify-content-center">
                  <!-- DoiT Sponsor -->
                  <div class="col-md-4 col-lg-3">
                    <div class="card shadow-sm h-100" style="min-height: 280px; border-color: #ffd700;">
                      <div class="card-body p-4 text-center d-flex flex-column">
                        <div class="mb-3 flex-grow-1 d-flex align-items-center justify-content-center">
                          <a href="https://doit.com" target="_blank">
                            <img src="/sponsors/doit.png"
                                 alt="DoiT International"
                                 class="img-fluid"
                                 style="max-height: 100px; max-width: 150px;">
                          </a>
                        </div>
                        <div class="mt-auto">
                          <h5 class="fw-bold mb-2">DoiT Cloud Intelligence™</h5>
                          <p class="small text-muted mb-0">
                            An intent-aware FinOps++ platform that goes far beyond finding idle servers.
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Digiusher Sponsor -->
                  <div class="col-md-4 col-lg-3">
                    <div class="card shadow-sm h-100" style="min-height: 280px; border-color: #ffd700;">
                      <div class="card-body p-4 text-center d-flex flex-column">
                        <div class="mb-3 flex-grow-1 d-flex align-items-center justify-content-center">
                          <a href="https://www.digiusher.com/contact/?utm_campaign=Cloud+Instances&utm_medium=Web&utm_source=LeanerCloud&utm_content=Cloud+Instances&utm_term=Sponsor" target="_blank">
                            <img src="/sponsors/digiusher.png"
                                 alt="Digiusher"
                                 class="img-fluid"
                                 style="max-height: 100px; max-width: 150px;">
                          </a>
                        </div>
                        <div class="mt-auto">
                          <h5 class="fw-bold mb-2">Digiusher</h5>
                          <p class="small text-muted mb-0">
                            Digital transformation and cloud consulting services.
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Gold CTA -->
                  <div class="col-md-4 col-lg-3">
                    <div class="card border-2 border-dashed h-100" style="min-height: 280px; border-color: #ffd700;">
                      <div class="card-body p-4 text-center d-flex flex-column justify-content-center">
                        <div class="mb-3">
                          <svg width="60" height="60" fill="#ffd700" viewBox="0 0 16 16">
                            <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"/>
                            <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
                          </svg>
                        </div>
                        <h5 class="fw-bold mb-2">Your Logo Here</h5>
                        <br/>
                        <a href="mailto:contact+sponsorship-opportunity@leanercloud.com.info?subject=Sponsorship%20Opportunity" class="btn btn-sm" style="background-color: #ffd700; border-color: #ffd700; color: #000;">
                          Become a Gold Sponsor
                        </a>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Silver Tier -->
          <div class="mb-5">
            <h2 class="h3 fw-bold mb-4 text-center">
              <span class="badge text-dark me-2" style="background-color: #c0c0c0;">Silver Sponsors</span>
            </h2>

            <div class="row justify-content-center">
              <div class="col-lg-8">
                <div class="row g-4 justify-content-center">
                  <!-- Silver CTA -->
                  <div class="col-md-6 col-lg-4">
                    <div class="card border-2 border-dashed h-100" style="min-height: 150px; border-color: #c0c0c0;">
                      <div class="card-body p-3 text-center d-flex flex-column justify-content-center">
                        <div class="mb-2">
                          <svg width="40" height="40" fill="#c0c0c0" viewBox="0 0 16 16">
                            <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"/>
                            <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
                          </svg>
                        </div>
                        <h6 class="fw-bold mb-2 small">Your Logo</h6>
                        <a href="mailto:contact+sponsorship-opportunity@leanercloud.com.info?subject=Sponsorship%20Opportunity" class="btn btn-sm small" style="background-color: #c0c0c0; border-color: #c0c0c0; color: #000; font-size: 11px;">
                          Sponsor
                        </a>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Bronze Tier -->
          <div class="mb-5">
            <h2 class="h3 fw-bold mb-4 text-center">
              <span class="badge text-white me-2" style="background-color: #cd7f32;">Bronze Sponsors</span>
            </h2>

            <div class="row justify-content-center">
              <div class="col-lg-4">
                <div class="row g-3 justify-content-center">
                  <!-- Bronze CTA -->
                  <div class="col-6 col-md-4">
                    <div class="card border-2 border-dashed h-100" style="min-height: 80px; border-color: #cd7f32;">
                      <div class="card-body p-2 text-center d-flex flex-column justify-content-center">
                        <div class="mb-1">
                          <svg width="24" height="24" fill="#cd7f32" viewBox="0 0 16 16">
                            <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"/>
                            <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
                          </svg>
                        </div>
                        <a href="mailto:contact+sponsorship-opportunity@leanercloud.com.info?subject=Sponsorship%20Opportunity" class="btn btn-sm" style="background-color: #cd7f32; border-color: #cd7f32; color: #fff; font-size: 10px; padding: 2px 6px;">
                          Sponsor
                        </a>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- One-off Sponsorships -->
          <div class="mb-5">
            <h2 class="h3 fw-bold mb-4 text-center">
              <span class="badge bg-info text-white me-2">Event & Conference Sponsors</span>
            </h2>

            <div class="row justify-content-center">
              <div class="col-lg-8">
                <div class="card bg-light">
                  <div class="card-body p-4">
                    <div class="text-center mb-4">
                      <h5 class="fw-bold mb-3">Thank you for supporting our community presence!</h5>
                      <p class="text-muted">
                        These organizations have sponsored our attendance at conferences, meetups, and community events, 
                        helping us connect with our users community worldwide.
                      </p>
                    </div>

                    <!-- FinOpsX 2025 Sponsors -->
                    <div class="mb-4">
                      <h6 class="fw-bold text-center mb-3">FinOpsX 2025 Conference Sponsors</h6>
                      <div class="row g-3">
                        <div class="col-md-4">
                          <div class="card border-info">
                            <div class="card-body p-3 text-center">
                              <h6 class="fw-bold mb-1">
                                <a href="https://datafy.io" target="_blank" class="text-decoration-none">Datafy.io</a>
                              </h6>
                              <small class="text-muted">Conference Attendance Costs</small>
                            </div>
                          </div>
                        </div>
                        <div class="col-md-4">
                          <div class="card border-info">
                            <div class="card-body p-3 text-center">
                              <h6 class="fw-bold mb-1">
                                <a href="https://infracost.io" target="_blank" class="text-decoration-none">Infracost.io</a>
                              </h6>
                              <small class="text-muted">Conference Attendance Costs</small>
                            </div>
                          </div>
                        </div>
                        <div class="col-md-4">
                          <div class="card border-info">
                            <div class="card-body p-3 text-center">
                              <h6 class="fw-bold mb-1">Anonymous Contributor</h6>
                              <small class="text-muted">Conference Attendance Costs</small>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <!-- Call to Action -->
                    <div class="text-center">
                      <a href="mailto:contact+sponsorship-opportunity@leanercloud.com.info?subject=Sponsorship%20Opportunity" class="btn btn-info">
                        Sponsor an Event
                      </a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Call to Action -->
          <div class="text-center mb-5">
            <div class="row justify-content-center">
              <div class="col-lg-8">
                <div class="card bg-primary text-white">
                  <div class="card-body p-5">
                    <h2 class="fw-bold mb-3">Become a Sponsor</h2>
                    <p class="lead mb-4">
                      Support the community and showcase your product by sponsoring Cloud-Instances.info.
                      Your sponsorship helps us maintain and improve this free resource used by thousands of
                      cloud engineers worldwide.
                    </p>

                    <div class="row g-3 mb-4">
                      <div class="col-md-4">
                        <div class="card bg-white text-dark h-100">
                          <div class="card-body p-3">
                            <h5 class="fw-bold" style="color: #ffd700;">Gold</h5>
                            <p class="small mb-0">Logo + Description + Link</p>
                          </div>
                        </div>
                      </div>
                      <div class="col-md-4">
                        <div class="card bg-white text-dark h-100">
                          <div class="card-body p-3">
                            <h5 class="fw-bold" style="color: #c0c0c0;">Silver</h5>
                            <p class="small mb-0">Logo + Link</p>
                          </div>
                        </div>
                      </div>
                      <div class="col-md-4">
                        <div class="card bg-white text-dark h-100">
                          <div class="card-body p-3">
                            <h5 class="fw-bold" style="color: #cd7f32;">Bronze</h5>
                            <p class="small mb-0">Name/Company + Link</p>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div class="d-flex flex-wrap justify-content-center gap-3">
                      <a href="mailto:contact+sponsorship-opportunity@leanercloud.com.info?subject=Sponsorship%20Opportunity" class="btn btn-light btn-lg">
                        <i class="material-icons me-2">email</i>
                        Contact Us
                      </a>
                      
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Community Recognition -->
          <div class="text-center">
            <h2 class="h4 fw-bold mb-3">Community Contributors</h2>
            <p class="text-muted mb-4">
              Special thanks to all the <a href="https://github.com/LeanerCloud/cloud-instances.info/contributors" target="_blank" class="text-decoration-none">
              GitHub contributors</a> who help improve and maintain this project through code contributions, 
              bug reports, and feature suggestions.
            </p>
            <div class="row justify-content-center">
              <div class="col-lg-6">
                <div class="card border-success">
                  <div class="card-body p-4">
                    <div class="d-flex align-items-center justify-content-center">
                      
                      <div>
                        <h5 class="fw-bold mb-1">Open Source Forever</h5>
                        <p class="text-muted mb-0">This project will always remain free and open source</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add some bottom padding to ensure all content is visible -->
    <div style="height: 100px;"></div>
```

### 2. render.py

This is the main rendering system that generates HTML pages from templates and JSON data. Key functions:
- `render()` - Main function to render HTML from templates and JSON data
- `sponsors_page()` - Specifically renders the sponsors page
- `about_page()` - Renders the about page
- `per_region_pricing()` - Splits pricing data by region
- `build_sitemap()` - Generates sitemap.xml
- Various helper functions for data processing

```python
import mako.template
import mako.lookup
import mako.exceptions
import io
import json
import datetime
import os
import copy
import yaml
import re

from detail_pages_ec2 import build_detail_pages_ec2
from detail_pages_rds import build_detail_pages_rds
from detail_pages_cache import build_detail_pages_cache
from detail_pages_opensearch import build_detail_pages_opensearch
from detail_pages_redshift import build_detail_pages_redshift


def network_sort(inst):
    perf = inst["network_performance"]
    network_rank = [
        "Very Low",
        "Low",
        "Low to Moderate",
        "Moderate",
        "High",
        "Up to 5 Gigabit",
        "Up to 10 Gigabit",
        "10 Gigabit",
        "12 Gigabit",
        "20 Gigabit",
        "Up to 25 Gigabit",
        "25 Gigabit",
        "50 Gigabit",
        "75 Gigabit",
        "100 Gigabit",
    ]
    try:
        sort = network_rank.index(perf)
    except ValueError:
        sort = len(network_rank)
    sort *= 2
    if inst.get("ebs_optimized"):
        sort += 1
    return sort


def add_cpu_detail(i):
    try:
        i["ECU_per_vcpu"] = i["ECU"] / i["vCPU"]
    except:
        # these will be instances with variable/burstable ECU
        i["ECU_per_vcpu"] = "unknown"

    try:
        if "vCPU" in i:
            # only EC2 uses vCPU
            i["memory_per_vcpu"] = round(i["memory"] / i["vCPU"], 2)
        else:
            i["memory_per_vcpu"] = round(float(i["memory"]) / float(i["vcpu"]), 2)
    except:
        # just to be safe...
        i["memory_per_vcpu"] = "unknown"

    if "physical_processor" in i:
        i["physical_processor"] = (i["physical_processor"] or "").replace("*", "")
        i["intel_avx"] = "Yes" if i["intel_avx"] else ""
        i["intel_avx2"] = "Yes" if i["intel_avx2"] else ""
        i["intel_avx512"] = "Yes" if i["intel_avx512"] else ""
        i["intel_turbo"] = "Yes" if i["intel_turbo"] else ""


def add_render_info(i):
    try:
        i["network_sort"] = network_sort(i)
    except KeyError:
        # This instance, probably from a non EC2 service, does not have traditional networking specs
        pass
    add_cpu_detail(i)


prices_dict = {}
prices_index = 0


def _compress_pricing(d):
    global prices_index

    for k, v in d.items():
        if k in prices_dict:
            nk = prices_dict[k]
        else:
            prices_dict[k] = nk = prices_index
            prices_index += 1

        if isinstance(v, dict):
            nv = dict(_compress_pricing(v))
        else:
            nv = v

        yield nk, nv


def compress_pricing(instances):
    global prices_index

    prices = {i["instance_type"]: i["pricing"] for i in instances}

    prices_dict.clear()
    prices_index = 0

    return json.dumps({"index": prices_dict, "data": dict(_compress_pricing(prices))})


def compress_instance_azs(instances):
    instance_type_region_availability_zones = {}
    for inst in instances:
        if "instance_type" in inst and "availability_zones" in inst:
            instance_type_region_availability_zones[inst["instance_type"]] = inst[
                "availability_zones"
            ]
    return json.dumps(instance_type_region_availability_zones)


def about_page(destination_file="www/about.html"):
    print("Rendering to %s..." % destination_file)
    lookup = mako.lookup.TemplateLookup(directories=["."])
    template = mako.template.Template(filename="in/about.html.mako", lookup=lookup)
    generated_at = datetime.datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC")
    os.makedirs(os.path.dirname(destination_file), exist_ok=True)
    with io.open(destination_file, "w+", encoding="utf-8") as fh:
        try:
            fh.write(template.render(generated_at=generated_at))
        except:
            print(mako.exceptions.text_error_template().render())
    return destination_file


def sponsors_page(destination_file="www/sponsors.html"):
    print("Rendering to %s..." % destination_file)
    lookup = mako.lookup.TemplateLookup(directories=["."])
    template = mako.template.Template(filename="in/sponsors.html.mako", lookup=lookup)
    generated_at = datetime.datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC")
    os.makedirs(os.path.dirname(destination_file), exist_ok=True)
    with io.open(destination_file, "w+", encoding="utf-8") as fh:
        try:
            rendered_content = template.render(generated_at=generated_at)
            # Convert to string regardless of input type
            if isinstance(rendered_content, (bytes, bytearray)):
                content_str = rendered_content.decode("utf-8")
            else:
                content_str = str(rendered_content)
            fh.write(content_str)
        except:
            print(mako.exceptions.text_error_template().render())
    return destination_file


def build_sitemap(sitemap):
    HOST = ""
    surls = ['<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">']
    for url in sitemap:
        surl = url.replace("www/", "")
        if "index" in surl:
            surl = surl.replace("index", "")
        surls.append("<url><loc>{}/{}</loc></url>".format(HOST, surl[0:-5]))
    surls.append("</urlset>")

    destination_file = "www/sitemap.xml"
    print("Rendering all URLs to %s..." % destination_file)
    with io.open(destination_file, "w+") as fp:
        fp.write("\n".join(surls))


def per_region_pricing(instances, data_file, all_regions):
    # This function splits instances.json into per-region files which are written to
    # disk and then can be loaded by the web app to reduce the amount of data that
    # needs to be sent to the client.

    init_pricing_json = ""
    init_instance_azs_json = ""

    outdir = data_file.replace("instances.json", "")

    instances_no_pricing = copy.deepcopy(instances)
    for i in instances_no_pricing:
        if "pricing" in i:
            del i["pricing"]
        if "availability_zones" in i:
            del i["availability_zones"]

    for r in all_regions:
        per_region_out = {}
        per_region_out = instances_no_pricing

        for i, inst in enumerate(instances):
            per_region_out[i]["pricing"] = {}
            per_region_out[i]["availability_zones"] = {}
            if r in inst["pricing"]:
                per_region_out[i]["pricing"][r] = instances[i]["pricing"][r]
            if "availability_zones" in inst and r in inst["availability_zones"]:
                per_region_out[i]["availability_zones"][r] = instances[i][
                    "availability_zones"
                ][r]

        pricing_out_file = "{}pricing_{}.json".format(outdir, r)
        azs_out_file = "{}instance_azs_{}.json".format(outdir, r)

        pricing_json = compress_pricing(per_region_out)
        instance_azs_json = compress_instance_azs(per_region_out)

        if r == "us-east-1":
            init_pricing_json = pricing_json
            init_instance_azs_json = instance_azs_json

        with open(pricing_out_file, "w+") as f:
            f.write(pricing_json)
        with open(azs_out_file, "w+") as f:
            f.write(instance_azs_json)

    return init_pricing_json, init_instance_azs_json


def regions_list(instances):
    regions = {}
    regions["main"] = {}
    regions["local_zone"] = {}
    regions["wavelength"] = {}

    for i in instances:
        for r in i["pricing"]:
            try:
                if "wl1" in r or "wl2" in r:
                    regions["wavelength"][r] = i["regions"][r]
                elif len(re.findall(r"\d+", r)) > 1:
                    regions["local_zone"][r] = i["regions"][r]
                else:
                    regions["main"][r] = i["regions"][r]
            except KeyError:
                print(
                    'ERROR: "regions" key not found in instances.json. Run scrape.py.'
                )

    return regions


def render(data_file, template_file, destination_file, detail_pages=True):
    """Build the HTML content from scraped data"""
    lookup = mako.lookup.TemplateLookup(directories=["."])
    template = mako.template.Template(filename=template_file, lookup=lookup)
    with open(data_file, "r") as f:
        instances = json.load(f)

    print("Loading data from %s..." % data_file)
    for i in instances:
        add_render_info(i)

    regions = regions_list(instances)

    sitemap = []
    if data_file == "www/instances.json":
        all_regions = regions["main"].copy()
        all_regions.update(regions["local_zone"])
        all_regions.update(regions["wavelength"])
        if detail_pages:
            sitemap.extend(build_detail_pages_ec2(instances, all_regions))
    elif data_file == "www/rds/instances.json":
        all_regions = regions["main"].copy()
        all_regions.update(regions["local_zone"])
        if detail_pages:
            sitemap.extend(build_detail_pages_rds(instances, all_regions))
    elif data_file == "www/cache/instances.json":
        all_regions = regions["main"].copy()
        if detail_pages:
            sitemap.extend(build_detail_pages_cache(instances, all_regions))
    elif data_file == "www/opensearch/instances.json":
        all_regions = regions["main"].copy()
        if detail_pages:
            sitemap.extend(build_detail_pages_opensearch(instances, all_regions))
    elif data_file == "www/redshift/instances.json":
        all_regions = regions["main"].copy()
        if detail_pages:
            sitemap.extend(build_detail_pages_redshift(instances, all_regions))

    generated_at = datetime.datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC")
    pricing_json, instance_azs_json = per_region_pricing(
        instances, data_file, all_regions
    )

    print("Rendering to %s..." % destination_file)
    os.makedirs(os.path.dirname(destination_file), exist_ok=True)
    with io.open(destination_file, "w+", encoding="utf-8") as fh:
        try:
            fh.write(
                template.render(
                    instances=instances,
                    regions=regions,
                    pricing_json=pricing_json,
                    generated_at=generated_at,
                    instance_azs_json=instance_azs_json,
                )
            )
            sitemap.append(destination_file)
        except:
            print(mako.exceptions.text_error_template().render())

    return sitemap


if __name__ == "__main__":
    sitemap = []
    sitemap.extend(render("www/instances.json", "in/index.html.mako", "www/index.html"))
    sitemap.extend(
        render("www/rds/instances.json", "in/rds.html.mako", "www/rds/index.html")
    )
    sitemap.extend(
        render(
            "www/cache/instances.json",
            "in/cache.html.mako",
            "www/cache/index.html",
        )
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
    sitemap.append(sponsors_page())
    build_sitemap(sitemap)
```

### 3. www/sponsors.html

This is the rendered HTML output of the sponsors page. It's the static HTML file served to users, generated from the Mako template above. The file contains the full rendered page with all sponsor information, tiers, and styling.

[Note: The full HTML content is preserved above in the file read output. It's the complete rendered version of the sponsors.html.mako template with all the navigation, sponsor cards, and footer elements.]