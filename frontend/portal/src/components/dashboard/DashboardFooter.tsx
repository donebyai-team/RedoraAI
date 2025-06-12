
import React from "react";
import { Heart, Github, Twitter } from "lucide-react";
import Link from "next/link";

const support_email = process.env.NEXT_PUBLIC_SUPPORT_EMAIL;

export function DashboardFooter() {
  return (
    <footer className="border-t border-primary/10 bg-background/95 py-4 px-4 md:px-6">
      <div className="container mx-auto flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
        <div className="flex items-center gap-2">
          <span>© {new Date().getFullYear()} Redora AI</span>
          <span className="flex items-center">
            Made with <Heart className="h-3 w-3 mx-1 text-destructive" /> by Team Redora
          </span>
        </div>

        <div className="flex items-center gap-4">
          <Link href="/terms" className="hover:text-primary transition-colors">
            Terms
          </Link>
          <Link href="/privacy" className="hover:text-primary transition-colors">
            Privacy
          </Link>
          <Link href={`mailto:${support_email}`} className="hover:text-primary transition-colors">
            Help
          </Link>
        </div>

        <div className="flex items-center gap-4">
          <a href="https://github.com" target="_blank" rel="noopener noreferrer" className="hover:text-primary transition-colors">
            <Github className="h-4 w-4" />
          </a>
          <a href="https://twitter.com" target="_blank" rel="noopener noreferrer" className="hover:text-primary transition-colors">
            <Twitter className="h-4 w-4" />
          </a>
        </div>
      </div>
    </footer>
  );
}
