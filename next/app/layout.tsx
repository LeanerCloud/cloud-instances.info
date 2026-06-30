import "./globals.css";
import { Inter } from "next/font/google";
import TopNav from "@/components/TopNav";
import Footer from "@/components/Footer";
import { Toaster } from "@/components/ui/sonner";
import { GoogleTagManager } from "@next/third-parties/google";
import Script from "next/script";

const inter = Inter({ subsets: ["latin"] });

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang="en">
            <head>
                <link rel="icon" href="/favicon.png" />
                <link
                    rel="sitemap"
                    type="application/xml"
                    title="Sitemap"
                    href="/sitemap_index.xml"
                />
                {process.env.NEXT_PUBLIC_GOOGLE_TAG_MANAGER_ID && (
                    <GoogleTagManager
                        gtmId={process.env.NEXT_PUBLIC_GOOGLE_TAG_MANAGER_ID}
                    />
                )}
                {process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID && (
                    <Script
                        src={
                            process.env.NEXT_PUBLIC_UMAMI_SCRIPT_URL ||
                            "https://stats.leanercloud.com/data.js"
                        }
                        data-website-id={
                            process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID
                        }
                        strategy="afterInteractive"
                    />
                )}
            </head>
            <body className={inter.className}>
                <TopNav />
                <Toaster theme="light" duration={2000} />
                {children}
                <Footer />
            </body>
        </html>
    );
}
