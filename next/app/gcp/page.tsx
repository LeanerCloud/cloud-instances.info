import { Metadata } from "next";
import { GCPClient } from "./GCPClient";
import { readdir, readFile } from "fs/promises";
import { join } from "path";
import * as YAML from "js-yaml";
import { decode } from "@msgpack/msgpack";

export const metadata: Metadata = {
  title: "GCP Compute Engine Comparison",
  description:
    "A free and easy-to-use tool for comparing GCP Compute Engine features and prices.",
};

export default async function GCPPage() {
  // Load regions data
  const regionsFile = await readFile(
    join(process.cwd(), "data/gcp-regions.yaml"),
    "utf8"
  );
  const regions = YAML.load(regionsFile);

  // Load instances data (will be generated from our Go scraper)
  let instances: any[] = [];
  let instanceCount = 0;
  
  try {
    const instancesFile = await readFile(
      join(process.cwd(), "data/gcp-instances.json"),
      "utf8"
    );
    instances = JSON.parse(instancesFile);
    instanceCount = instances.length;
  } catch (error) {
    console.warn("GCP instances data not found, using empty dataset");
  }

  return (
    <>
      <link
        rel="preload"
        href="/data/gcp-instances.msgpack.xz"
        as="fetch"
        crossOrigin="anonymous"
      />
      <link
        rel="preload"
        href="/data/gcp-regions.yaml"
        as="fetch"
        crossOrigin="anonymous"
      />
      <GCPClient
        columnAtomKey="gcp"
        regions={regions}
        instances={instances}
        instanceCount={instanceCount}
      />
    </>
  );
}