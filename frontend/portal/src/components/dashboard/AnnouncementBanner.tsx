import { useState } from "react";
import { X, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface AnnouncementBannerProps {
    message: string;
}

export function AnnouncementBanner({ message }: AnnouncementBannerProps) {
    const [isVisible, setIsVisible] = useState(true);

    if (!isVisible) {
        return null;
    }

    return (
        <Alert className="rounded-none border-l-0 border-r-0 border-t-0 bg-gradient-to-r from-blue-600 to-purple-600 border-blue-500 text-white shadow-md py-2 px-4">
            <AlertDescription className="flex items-center justify-between w-full">
                {/* Left side: icon + message */}
                <div className="flex items-center gap-2">
                    {/* <Sparkles className="h-4 w-4 text-yellow-300" /> */}
                    <span className="text-sm font-medium leading-tight">{message}</span>
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
