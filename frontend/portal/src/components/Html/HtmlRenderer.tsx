import React, { useEffect, useMemo, useRef, useState } from "react";
import he from "he";
import ReactMarkdown from 'react-markdown';
import { ChevronUp, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleTrigger } from "@radix-ui/react-collapsible";

interface HtmlRendererProps {
  htmlString: string;
}

interface MarkdownRendererProps {
  data: string;
}

const HtmlTitleRenderer: React.FC<HtmlRendererProps> = ({ htmlString }) => {
  const decodedHtml = he.decode(htmlString);

  return (<div style={{ all: "revert" }} dangerouslySetInnerHTML={{ __html: decodedHtml }} />);
};

const HtmlBodyRenderer = ({ htmlString }: { htmlString: string }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const decodedHtml = he.decode(htmlString);

  useEffect(() => {
    if (!isExpanded) return;

    const iframe = iframeRef.current;
    if (!iframe) return;

    const doc = iframe.contentDocument || iframe.contentWindow?.document;
    if (!doc) return;

    doc.open();
    doc.write(`
      <html>
        <head>
          <style>
            html, body {
              margin: 0;
              padding: 0;
              overflow: hidden;
              color: rgb(100, 116, 139);
              font-family: system-ui, sans-serif;
              font-size: 14px;
              line-height: 1.5;
            }
            a {
              color: #3b82f6;
              text-decoration: underline;
            }
          </style>
        </head>
        <body>
          ${decodedHtml}
        </body>
      </html>
    `);
    doc.close();

    const resizeIframe = () => {
      if (iframe && doc.body) {
        iframe.style.height = doc.body.scrollHeight + "px";
      }

      const links = doc.querySelectorAll("a");
      links.forEach((link) => {
        link.setAttribute("target", "_blank");
        link.setAttribute("rel", "noopener noreferrer");
      });
    };

    setTimeout(resizeIframe, 0);
  }, [decodedHtml, isExpanded]);

  // Optional: force rerender on content change
  const iframeKey = useMemo(() => {
    return `iframe-${decodedHtml.length}-${Date.now()}`;
  }, [decodedHtml]);

  return (
    <div className="space-y-2">
      {!isExpanded ? (
        <div className="text-sm text-gray-800 line-clamp-2">{decodedHtml.replace(/<\/?[^>]+(>|$)/g, "")}</div>
      ) : (
        <iframe
          key={iframeKey}
          ref={iframeRef}
          style={{
            width: "100%",
            border: "none",
            overflow: "hidden",
            height: "1px", // auto-updated on load
          }}
          scrolling="no"
          sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
          title="HTML Preview"
        />
      )}

      <Button
        variant="ghost"
        size="sm"
        className="h-auto p-0 text-blue-600 hover:text-blue-800"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <span className="flex items-center gap-1 text-xs">
          {isExpanded ? (
            <>
              Read less <ChevronUp className="h-3 w-3" />
            </>
          ) : (
            <>
              Read more <ChevronDown className="h-3 w-3" />
            </>
          )}
        </span>
      </Button>
    </div>
  );
};


const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ data }) => {
  const processedData = data.replace(/\n/g, "\n\n");
  return (
    <ReactMarkdown
      components={{
        p: ({ ...props }) => (
          <p style={{ all: "revert", color: "rgb(100, 116, 139)" }} {...props} />
        ),
      }}
    >
      {processedData}
    </ReactMarkdown>
  );
}

const CollapsibleText = ({ text }: { text: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  const shouldCollapse = text.length > 100;

  const renderMarkdown = (content: string) => (
    <ReactMarkdown
      children={content}
      components={{
        p: ({ children }) => <p className="text-sm text-gray-700">{children}</p>,
      }}
    />
  );

  if (!shouldCollapse) {
    return renderMarkdown(text);
  }

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <div className="space-y-2">
        {isOpen
          ? renderMarkdown(text)
          : renderMarkdown(text.substring(0, 100) + '...')}
        <CollapsibleTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-auto p-0 text-blue-600 hover:text-blue-800"
          >
            <span className="flex items-center gap-1 text-xs">
              {isOpen ? (
                <>
                  Read less <ChevronUp className="h-3 w-3" />
                </>
              ) : (
                <>
                  Read more <ChevronDown className="h-3 w-3" />
                </>
              )}
            </span>
          </Button>
        </CollapsibleTrigger>
      </div>
    </Collapsible>
  );
};


export { HtmlTitleRenderer, HtmlBodyRenderer, MarkdownRenderer, CollapsibleText };
