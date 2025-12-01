import { Metadata } from "next";
import { GCPClient } from "./GCPClient";
import { readFile } from "fs/promises";
import Head from "next/head";
import { PIPELINE_SIZE } from "@/utils/handleCompressedFile";
import * as YAML from "js-yaml";
import { decode } from "@msgpack/msgpack";
import { GCPInstance } from "@/utils/colunnData/gcp";

export const metadata: Metadata = {
  title: "GCP Compute Engine Comparison",
  description:
    "A free and easy-to-use tool for comparing GCP Compute Engine features and prices.",
};

export default async function GCPPage() {
  // Load regions data
  const regionsFile = await readFile("./public/data/gcp-regions.yaml", "utf8");
  const regions = YAML.load(regionsFile);

  const data = await readFile("./public/first-100-gcp-instances.msgpack");
  const compressedInstances = decode(data) as [string[], ...GCPInstance[]];

  const instanceCount = Number(
    await readFile("./public/gcp-instance-count.txt"),
  );
  const instancesHash = await readFile(
    "./public/gcp-instances-hash.txt",
    "utf-8",
  );

  return (
    <>
      <Head>
        {Array.from({ length: PIPELINE_SIZE }).map((_, i) => (
          <link
            key={i}
            rel="preload"
            href={`/remaining-gcp-instances-p${i}.msgpack.xz?cache=${instancesHash}`}
            as="fetch"
            fetchPriority="high"
          />
        ))}
      </Head>
      <GCPClient
        columnAtomKey="gcp"
        regions={regions}
        instanceCount={instanceCount}
        compressedDataPathTemplate={`/remaining-gcp-instances-p{}.msgpack.xz?cache=${instancesHash}`}
        compressedInstances={compressedInstances}
      />
    </>
  );
}