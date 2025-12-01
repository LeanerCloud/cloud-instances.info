import { Metadata } from "next";
import { OCIClient } from "./OCIClient";
import { readFile } from "fs/promises";
import Head from "next/head";
import { PIPELINE_SIZE } from "@/utils/handleCompressedFile";
import * as YAML from "js-yaml";
import { decode } from "@msgpack/msgpack";
import { OCIInstance } from "@/utils/colunnData/oci";

export const metadata: Metadata = {
  title: "OCI Compute Instance Comparison",
  description:
    "A free and easy-to-use tool for comparing Oracle Cloud Infrastructure Compute instance features and prices.",
};

export default async function OCIPage() {
  // Load regions data
  const regionsFile = await readFile("./public/data/oci-regions.yaml", "utf8");
  const regions = YAML.load(regionsFile);

  const data = await readFile("./public/first-100-oci-instances.msgpack");
  const compressedInstances = decode(data) as [string[], ...OCIInstance[]];

  const instanceCount = Number(
    await readFile("./public/oci-instance-count.txt"),
  );
  const instancesHash = await readFile(
    "./public/oci-instances-hash.txt",
    "utf-8",
  );

  return (
    <>
      <Head>
        {Array.from({ length: PIPELINE_SIZE }).map((_, i) => (
          <link
            key={i}
            rel="preload"
            href={`/remaining-oci-instances-p${i}.msgpack.xz?cache=${instancesHash}`}
            as="fetch"
            fetchPriority="high"
          />
        ))}
      </Head>
      <OCIClient
        columnAtomKey="oci"
        regions={regions}
        instanceCount={instanceCount}
        compressedDataPathTemplate={`/remaining-oci-instances-p{}.msgpack.xz?cache=${instancesHash}`}
        compressedInstances={compressedInstances}
      />
    </>
  );
}