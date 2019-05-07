class TfvarsAnnotations < Formula
  desc "Update values in terraform.tfvars using annotations"
  homepage "https://github.com/antonbabenko/tfvars-annotations"

  # Update these when a new version is released
  url "https://github.com/antonbabenko/tfvars-annotations/archive/v0.0.3.tar.gz"
  sha256 "d2f2e6afbba4b4901cbb6bb04713fc17769df9f07eb7675449ea011e4ac8cf3e"

  head "https://github.com/antonbabenko/tfvars-annotations.git"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath

    # Move the contents of the repo (which are currently in the buildpath) into
    # a go-style subdir, so we can build it without spewing deps everywhere.
    app_path = buildpath/"src/github.com/antonbabenko/tfvars-annotations"
    app_path.install Dir["*"]

    # Fetch the deps (into our temporary gopath) and build
    cd "src/github.com/antonbabenko/tfvars-annotations" do
      system "go", "get"
      system "go", "build", "-ldflags", "-X main.buildVersion='#{version}'"
    end

    # Install the resulting binary
    bin.install "bin/tfvars-annotations"
  end

  test do
    system "#{bin}/tfvars-annotations", "-version"
  end
end
