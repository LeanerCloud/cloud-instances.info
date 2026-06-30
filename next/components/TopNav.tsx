"use client";

import TranslationFriendlyLink from "@/components/TranslationFriendlyLink";
import { buttonVariants } from "@/components/ui/button";
import { usePathname } from "next/navigation";
import { useEffect, useRef } from "react";
import { translationToolDetected } from "@/state";

const navItems = [
    {
        label: "AWS",
        href: "/",
        children: [
            {
                label: "EC2",
                href: "/",
            },
            {
                label: "RDS",
                href: "/rds",
            },
            {
                label: "ElastiCache",
                href: "/cache",
            },
            {
                label: "Redshift",
                href: "/redshift",
            },
            {
                label: "OpenSearch",
                href: "/opensearch",
            },
        ],
    },
    {
        label: "Azure",
        href: "/azure",
    },
    {
        label: "GCP",
        href: "/gcp",
    },
];

function TranslationToolDetector({
    className,
    text,
}: {
    className: string;
    text: string;
}) {
    const spanRef = useRef<HTMLSpanElement>(null);

    useEffect(() => {
        const span = spanRef.current;
        if (!span) return;

        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                const textContent = mutation.target.textContent;
                if (textContent !== text) {
                    translationToolDetected.set(true);
                    observer.disconnect();
                }
            });
        });
        observer.observe(span, { childList: true });
        return () => observer.disconnect();
    }, [text, spanRef.current, translationToolDetected]);

    return (
        <span ref={spanRef} className={className}>
            {text}
        </span>
    );
}

type NavItemProps = {
    item: {
        label: string;
        href: string;
        children?: {
            label: string;
            href: string;
        }[];
    };
    currentPath: string;
};

function NavItem({ item, currentPath }: NavItemProps) {
    let anySelected = false;
    const children =
        item.children &&
        item.children.map((child) => {
            const selected =
                currentPath === child.href ||
                currentPath.includes(child.label.toLowerCase());
            if (selected) anySelected = true;
            return (
                <TranslationFriendlyLink
                    aria-current={selected}
                    className={`font-normal text-sm px-2 py-1 pb-2 rounded rounded-b-none ${
                        selected
                            ? "bg-white text-black font-semibold"
                            : "text-white"
                    }`}
                    key={child.label}
                    href={child.href}
                >
                    {child.label}
                </TranslationFriendlyLink>
            );
        });
    if (
        !anySelected &&
        currentPath.startsWith(item.href) &&
        item.href !== "/"
    ) {
        anySelected = true;
    }

    return (
        <div
            className="flex items-center justify-start gap-4 relative top-1.5 ml-2"
            key={item.label}
        >
            <TranslationFriendlyLink
                className="font-medium text-gray-4 text-sm"
                href={item.href}
                aria-current={anySelected}
            >
                <span
                    className={
                        anySelected ? "font-bold text-white" : "text-white"
                    }
                >
                    {item.label}
                </span>
            </TranslationFriendlyLink>
            {children && (
                <div className="flex items-center justify-start gap-4 rounded-md rounded-b-none bg-black/30 not-lg:hidden p-1 pb-0">
                    {children}
                </div>
            )}
        </div>
    );
}

export default function TopNav() {
    const currentPath = usePathname();

    return (
        <nav className="flex items-center justify-between bg-purple-brand h-[3rem] py-2 px-4 dark:bg-purple-dark">
            <div className="flex items-center justify-start gap-4">
                <TranslationFriendlyLink
                    href="/"
                    className="font-medium text-gray-4"
                >
                    <div className="flex items-center justify-start gap-2">
                        <div className="flex flex-col">
                            <span className="font-semibold text-white leading-5">
                                Cloud Instances
                            </span>
                            <TranslationToolDetector
                                className="text-xs italic text-gray-3"
                                text="vendor-neutral fork maintained by LeanerCloud"
                            />
                        </div>
                    </div>
                </TranslationFriendlyLink>
                {navItems.map((item) => (
                    <NavItem
                        key={item.label}
                        item={item}
                        currentPath={currentPath}
                    />
                ))}
            </div>
            <div className="flex items-center justify-end gap-4 not-md:hidden overflow-hidden">
                <TranslationFriendlyLink
                    href="/sponsors"
                    className={buttonVariants({
                        variant: "outline",
                        size: "sm",
                    })}
                >
                    Sponsors
                </TranslationFriendlyLink>
                <TranslationFriendlyLink
                    href="https://join.slack.com/t/leanercloud/shared_invite/zt-xodcoi9j-1IcxNozXx1OW0gh_N08sjg"
                    target="_blank"
                    className={buttonVariants({
                        variant: "outline",
                        size: "sm",
                    })}
                >
                    Slack
                </TranslationFriendlyLink>
                <TranslationFriendlyLink
                    href="https://github.com/LeanerCloud/cloud-instances.info"
                    className={buttonVariants({
                        variant: "outline",
                        size: "sm",
                    })}
                >
                    <svg
                        width="18"
                        height="18"
                        viewBox="0 0 18 18"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <g clipPath="url(#clip0_3_29)">
                            <path
                                d="M9 0C13.9725 0 18 4.0275 18 9C17.9995 10.8857 17.4077 12.7238 16.3078 14.2556C15.2079 15.7873 13.6554 16.9356 11.8688 17.5387C11.4188 17.6287 11.25 17.3475 11.25 17.1112C11.25 16.8075 11.2613 15.84 11.2613 14.6363C11.2613 13.7925 10.98 13.2525 10.6538 12.9712C12.6563 12.7463 14.76 11.9812 14.76 8.5275C14.76 7.5375 14.4113 6.73875 13.8375 6.10875C13.9275 5.88375 14.2425 4.96125 13.7475 3.72375C13.7475 3.72375 12.9938 3.47625 11.2725 4.64625C10.5525 4.44375 9.7875 4.3425 9.0225 4.3425C8.2575 4.3425 7.4925 4.44375 6.7725 4.64625C5.05125 3.4875 4.2975 3.72375 4.2975 3.72375C3.8025 4.96125 4.1175 5.88375 4.2075 6.10875C3.63375 6.73875 3.285 7.54875 3.285 8.5275C3.285 11.97 5.3775 12.7463 7.38 12.9712C7.12125 13.1962 6.885 13.59 6.80625 14.175C6.28875 14.4113 4.995 14.7937 4.185 13.4325C4.01625 13.1625 3.51 12.4987 2.80125 12.51C2.0475 12.5212 2.4975 12.9375 2.8125 13.1062C3.195 13.32 3.63375 14.1187 3.735 14.3775C3.915 14.8837 4.5 15.8513 6.76125 15.435C6.76125 16.1888 6.7725 16.8975 6.7725 17.1112C6.7725 17.3475 6.60375 17.6175 6.15375 17.5387C4.36122 16.9421 2.80208 15.7961 1.6975 14.2635C0.592928 12.7308 -0.000990539 10.8892 1.2401e-06 9C1.2401e-06 4.0275 4.0275 0 9 0Z"
                                fill="currentColor"
                            />
                        </g>
                        <defs>
                            <clipPath id="clip0_3_29">
                                <rect width="18" height="18" fill="white" />
                            </clipPath>
                        </defs>
                    </svg>
                    Star
                </TranslationFriendlyLink>
            </div>
        </nav>
    );
}
