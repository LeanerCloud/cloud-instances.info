import { Metadata } from "next";
import { OCIClient } from "./OCIClient";
import { readdir, readFile } from "fs/promises";
import { join } from "path";
import * as YAML from "js-yaml";
import { decode } from "@msgpack/msgpack";

export const metadata: Metadata = {
  title: "OCI Compute Instance Comparison",
  description:
    "A free and easy-to-use tool for comparing Oracle Cloud Infrastructure Compute instance features and prices.",
};

export default async function OCIPage() {
  // Load regions data
  const regionsFile = await readFile(
    join(process.cwd(), "data/oci-regions.yaml"),
    "utf8"
  );
  const regions = YAML.load(regionsFile);

  // Load instances data (will be generated from our Go scraper)
  let instances: any[] = [];
  let instanceCount = 0;
  
  try {
    const instancesFile = await readFile(
      join(process.cwd(), "data/oci-instances.json"),
      "utf8"
    );
    instances = JSON.parse(instancesFile);
    instanceCount = instances.length;
  } catch (error) {
    console.warn("OCI instances data not found, using empty dataset");
  }

  return (
    <>
      <link
        rel="preload"
        href="/data/oci-instances.msgpack.xz"
        as="fetch"
        crossOrigin="anonymous"
      />
      <link
        rel="preload"
        href="/data/oci-regions.yaml"
        as="fetch"
        crossOrigin="anonymous"
      />
      <OCIClient
        columnAtomKey="oci"
        regions={regions}
        instances={instances}
        instanceCount={instanceCount}
      />
    </>
  );
}