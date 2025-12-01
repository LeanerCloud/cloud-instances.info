import { Metadata } from "next";

export const metadata: Metadata = {
    title: "Sponsors - Cloud-Instances.info",
    description:
        "Supporting the development and maintenance of Cloud-Instances.info - Thank you to our sponsors and supporters!",
};

const goldSponsors = [
    {
        name: "DoiT",
        url: "https://doit.com",
        logo: "/sponsors/doit.png",
        description:
            "An intent-aware FinOps++ platform that goes far beyond finding idle servers.",
    },
    {
        name: "Digiusher",
        url: "https://www.digiusher.com/contact/?utm_campaign=Cloud+Instances&utm_medium=Web&utm_source=LeanerCloud&utm_content=Cloud+Instances&utm_term=Sponsor",
        logo: "/sponsors/digiusher.png",
        description:
            "Cloud consulting and managed services for AWS, Azure, and GCP.",
    },
];

export default function SponsorsPage() {
    return (
        <div className="min-h-screen bg-gray-50 py-12 px-4">
            <div className="max-w-4xl mx-auto">
                {/* Header */}
                <div className="text-center mb-12">
                    <h1 className="text-4xl font-bold mb-4">Why Sponsor Us?</h1>
                    <div className="text-gray-600 space-y-4 max-w-2xl mx-auto text-left">
                        <p>
                            The open source software powering Cloud-Instances.info
                            has been built for many years by a few dedicated
                            volunteers working on it in their limited spare time.
                            The main author eventually gave up and sold it to
                            Vantage, becoming a very important marketing asset.
                        </p>
                        <p>
                            After a few years of active development by Vantage, it
                            became neglected and development happened only in a
                            private branch for a while.
                        </p>
                        <p>
                            @cristim, a former co-maintainer before the Vantage
                            acquisition decided to fork the open source code and
                            now trying to bring it back into a community driven
                            project, actively maintained and developed in the open,
                            with the main page free of marketing messaging since
                            most users use it for work.
                        </p>
                        <p>
                            But at the same time we want to fund a few people to
                            work on it consistently, and for that we accept
                            sponsorships of all kinds.
                        </p>
                        <p className="font-semibold">
                            We really appreciate any help to keep this tool
                            actively maintained and vendor-neutral and are very
                            thankful to our current sponsors!
                        </p>
                    </div>
                </div>

                {/* Gold Sponsors */}
                <div className="mb-12">
                    <h2 className="text-2xl font-bold text-center mb-8">
                        <span className="inline-block px-4 py-1 rounded bg-yellow-400 text-yellow-900">
                            Gold Sponsors
                        </span>
                    </h2>
                    <div className="grid md:grid-cols-2 gap-6 justify-items-center">
                        {goldSponsors.map((sponsor) => (
                            <a
                                key={sponsor.name}
                                href={sponsor.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="block w-full max-w-sm"
                            >
                                <div className="bg-white rounded-lg shadow-md border-2 border-yellow-400 p-6 h-full hover:shadow-lg transition-shadow">
                                    <div className="flex items-center justify-center h-24 mb-4">
                                        <img
                                            src={sponsor.logo}
                                            alt={sponsor.name}
                                            className="max-h-full max-w-[150px] object-contain"
                                        />
                                    </div>
                                    <h3 className="text-xl font-bold text-center mb-2">
                                        {sponsor.name}
                                    </h3>
                                    <p className="text-gray-600 text-sm text-center">
                                        {sponsor.description}
                                    </p>
                                </div>
                            </a>
                        ))}
                    </div>
                </div>

                {/* Become a Sponsor */}
                <div className="bg-white rounded-lg shadow-md p-8 text-center">
                    <h2 className="text-2xl font-bold mb-4">
                        Become a Sponsor
                    </h2>
                    <p className="text-gray-600 mb-6">
                        Interested in supporting Cloud-Instances.info? We offer
                        various sponsorship tiers with different benefits.
                    </p>
                    <div className="space-y-4">
                        <a
                            href="https://github.com/sponsors/LeanerCloud"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-block bg-purple-600 text-white px-6 py-3 rounded-lg font-semibold hover:bg-purple-700 transition-colors"
                        >
                            Sponsor on GitHub
                        </a>
                        <p className="text-sm text-gray-500">
                            Or contact us at{" "}
                            <a
                                href="mailto:sponsors@leanercloud.com"
                                className="text-purple-600 hover:underline"
                            >
                                sponsors@leanercloud.com
                            </a>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
}
