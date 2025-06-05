import { useState } from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import Link from "next/link";

interface AnnouncementBannerProps {
    message: string;
    buttonText?: string;
    buttonHref?: string;
}

export function AnnouncementBanner({
    message,
    buttonText,
    buttonHref,
}: AnnouncementBannerProps) {
    const [isVisible, setIsVisible] = useState(true);

    if (!isVisible) {
        return null;
    }

    return (
        <Alert className="rounded-none border-l-0 border-r-0 border-t-0 bg-gradient-to-r from-blue-600 to-purple-600 border-blue-500 text-white shadow-md py-2 px-4">
            <AlertDescription className="flex items-center justify-between w-full">
                {/* Left side: message + optional button */}
                <div className="flex items-center gap-3">
                    <span className="text-sm font-medium leading-tight">{message}</span>

                    {buttonText && buttonHref && (
                        <Link href={buttonHref}>
                            <Button
                                variant="secondary"
                                size="sm"
                                className="text-xs text-blue-700 bg-white hover:bg-gray-100 h-6 px-2 py-0"
                            >
                                {buttonText}
                            </Button>

                        </Link>
                    )}
                </div>

                {/* Close button */}
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setIsVisible(false)}
                    className="h-6 w-6 text-white hover:text-gray-200 hover:bg-white/20 transition-colors"
                >
                    <X className="h-4 w-4" />
                </Button>
            </AlertDescription>
        </Alert>
    );
}
