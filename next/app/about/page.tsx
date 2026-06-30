import { Metadata } from "next";

export const metadata: Metadata = {
    title: "About - Cloud-Instances.info",
    description:
        "About Cloud-Instances.info - a community-driven, vendor-neutral cloud instance comparison tool.",
};

export default function AboutPage() {
    return (
        <div className="min-h-screen bg-gray-50 py-12 px-4">
            <div className="max-w-3xl mx-auto space-y-10">
                <h1 className="text-4xl font-bold text-center">About</h1>

                <section>
                    <h2 className="text-2xl font-bold mb-3">Why?</h2>
                    <p className="text-gray-600">
                        Because it&apos;s frustrating to compare instances using
                        Amazon&apos;s own instance type, pricing, and other
                        pages.
                    </p>
                </section>

                <section>
                    <h2 className="text-2xl font-bold mb-3">Who?</h2>
                    <p className="text-gray-600">
                        It was started by{" "}
                        <a
                            href="https://github.com/powdahound"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-purple-600 hover:underline"
                        >
                            @powdahound
                        </a>
                        , contributed to by many, and is now managed and
                        maintained by Cristian Magherusan-Stanciu, developed in
                        the open together with the wider Open Source community,
                        and awaits your improvements on{" "}
                        <a
                            href="https://github.com/LeanerCloud/cloud-instances.info"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-purple-600 hover:underline"
                        >
                            GitHub
                        </a>
                        . In the development of Detail Pages, we used designs
                        from cloudhw.info with permission from Joshua Powers.
                    </p>
                </section>

                <section>
                    <h2 className="text-2xl font-bold mb-3">Warning</h2>
                    <p className="text-gray-600">
                        This site is not maintained by or affiliated with
                        Amazon. The data shown is not guaranteed to be accurate
                        or current. Please{" "}
                        <a
                            href="https://github.com/LeanerCloud/cloud-instances.info/issues/new/"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-purple-600 hover:underline"
                        >
                            report issues
                        </a>{" "}
                        you see.
                    </p>
                </section>
            </div>
        </div>
    );
}
