"use client";

// Sponsor data - can be moved to a JSON file or API later
const sponsors = [
    {
        name: "DoiT",
        url: "https://doit.com",
        description: "Cloud Intelligence Platform",
    },
    {
        name: "Digiusher",
        url: "https://www.digiusher.com/contact/?utm_campaign=Cloud+Instances&utm_medium=Web&utm_source=LeanerCloud",
        description: "Cloud Consulting",
    },
];

const style = {
    color: "white",
    background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
};

export default function Advert(_props: {
    gpu?: boolean;
    instanceGroup?: string;
    marketingData?: unknown;
}) {
    if (process.env.NEXT_PUBLIC_REMOVE_ADVERTS === "1") return null;

    return (
        <div className="h-[2.5em] overflow-hidden" style={style}>
            <div className="flex items-center justify-center h-full gap-4 animate-marquee whitespace-nowrap">
                <span className="text-sm opacity-75 flex-shrink-0">
                    Sponsors:
                </span>
                {sponsors.map((sponsor, index) => (
                    <a
                        key={sponsor.name}
                        href={sponsor.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 flex-shrink-0 hover:opacity-80 transition-opacity"
                    >
                        <span className="font-bold">{sponsor.name}</span>
                        <span className="text-sm opacity-75">
                            ({sponsor.description})
                        </span>
                        {index < sponsors.length - 1 && (
                            <span className="mx-2 opacity-50">|</span>
                        )}
                    </a>
                ))}
                <span className="text-sm opacity-75 flex-shrink-0 ml-4">
                    <a
                        href="/sponsors"
                        className="underline hover:no-underline"
                    >
                        Become a sponsor
                    </a>
                </span>
            </div>
        </div>
    );
}
