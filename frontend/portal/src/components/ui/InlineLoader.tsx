import { Loader2 } from "lucide-react";
import React from "react";

interface InlineLoaderProps {
    isVisible?: boolean;
    size?: number;
    className?: string;
    containerClassName?: string;
}

export const InlineLoader: React.FC<InlineLoaderProps> = ({
    isVisible = true,
    size = 24,
    className = "",
    containerClassName = "flex justify-center my-4",
}) => {
    if (!isVisible) return null;

    return (
        <div className={containerClassName}>
            <Loader2 className={`animate-spin ${className}`} size={size} />
        </div>
    );
};
