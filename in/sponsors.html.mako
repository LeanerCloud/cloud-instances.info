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