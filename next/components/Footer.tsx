export default function Footer() {
    return (
        <div className="border-t border-gray-3 h-[3rem] sticky bottom-0 text-xs text-gray-2 bg-background">
            <div className="flex items-center justify-between h-full px-2">
                <div className="flex items-center gap-3">
                    <div className="hidden md:block">
                        Updated {new Date().toLocaleString()}
                    </div>
                </div>
                <div className="hidden md:block">
                    Cloud-Instances.info - Easy Cloud Instance Comparison
                </div>
                <div className="flex items-center gap-3">
                    <span>
                        vendor-neutral fork maintained by{" "}
                        <a
                            target="_blank"
                            href="https://leanercloud.com"
                            className="text-purple-brand text-underline hover:text-purple-0"
                        >
                            LeanerCloud
                        </a>
                    </span>
                    <a
                        href="/about"
                        className="text-purple-brand text-underline hover:text-purple-0"
                    >
                        About
                    </a>
                    <a
                        href="https://github.com/LeanerCloud/cloud-instances.info/issues/new/"
                        target="_blank"
                        className="text-purple-brand text-underline hover:text-purple-0"
                    >
                        Report issues
                    </a>
                </div>
            </div>
        </div>
    );
}
