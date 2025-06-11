
"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CreditCard, Calendar, Star, Crown, Zap, Check } from "lucide-react";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
// import { Button } from "@/components/ui/button";
// import { toast } from "@/hooks/use-toast";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { SubscriptionStatus } from "@doota/pb/doota/core/v1/core_pb";
import { Timestamp } from "@bufbuild/protobuf/wkt";
import { SubscriptionPlan } from "@doota/ui-core/context/AuthContext";
import { formatTimestampToReadableDate } from "@/utils/format";

interface UserSubscription {
    plan: SubscriptionPlan;
    expiryDate: Timestamp | undefined;
    isActive: boolean;
}

interface PlanInfo {
    name: "FREE" | "Founder" | "Pro";
    price: string;
    interval: string;
    icon: React.ReactNode;
    features: string[];
    popular?: boolean;
    description: string;
}

const plans: PlanInfo[] = [
    {
        name: "FREE",
        price: "$0",
        interval: "7-day trial",
        description: "7-day trial of Founder plan",
        icon: <Star className="h-5 w-5" />,
        features: [
            "5 Tracked Keywords",
            "5 Monitored Subreddits",
            "Up to 25 highly relevant posts per day",
            "Comment and DM suggestions",
            "Up to 25 automated comments per day",
            "Up to 25 automated DMs per day",
            "Real-time Alerts (In-app only)",
            "Basic Analytics"
        ]
    },
    {
        name: "Founder",
        price: "$39",
        interval: "per month",
        description: "For indie founders doing cold outreach",
        icon: <Crown className="h-5 w-5" />,
        features: [
            "5 Tracked Keywords",
            "5 Monitored Subreddits",
            "Up to 25 highly relevant posts per day",
            "Comment and DM suggestions",
            "Up to 25 automated comments per day",
            "Up to 25 automated DMs per day",
            "Real-time Alerts (In-app only)",
            "Basic Analytics"
        ],
        popular: true
    },
    {
        name: "Pro",
        price: "$99",
        interval: "per month",
        description: "For sales & growth teams",
        icon: <Zap className="h-5 w-5" />,
        features: [
            "20 Tracked Keywords",
            "20 Monitored Subreddits",
            "Up to 50 highly relevant posts per day",
            "Comment and DM suggestions",
            "Up to 50 automated comments per day",
            "Up to 50 automated DMs per day",
            "Up to 4 Campaigns/Projects",
            "Real-time Alerts (Slack/Email)",
            "CRM/Export Integration",
            "Priority Support (Email)"
        ]
    }
];

export default function Billing() {
    const { planDetails } = useAuth();
    const [subscription, setSubscription] = useState<UserSubscription>({
        plan: SubscriptionPlan.FREE,
        expiryDate: undefined,
        isActive: false,
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Simulate fetching subscription data
        const fetchSubscriptionData = async () => {
            try {
                setTimeout(() => {
                    setSubscription({
                        plan: planDetails?.planName,
                        expiryDate: planDetails.subscription?.expiresAt,
                        isActive: planDetails.subscription?.status === SubscriptionStatus.ACTIVE,
                    });
                    setLoading(false);
                }, 1000);
            } catch (error) {
                console.error("Error fetching subscription:", error);
                setLoading(false);
            }
        };

        fetchSubscriptionData();
    }, [planDetails]);

    if (loading) {
        return (
            <div className="min-h-screen bg-gradient-to-b from-background to-secondary/20">
                <DashboardHeader />
                <main className="container mx-auto px-4 py-6 md:px-6">
                    <div className="animate-pulse">
                        <div className="h-8 bg-gray-200 rounded w-1/4 mb-2"></div>
                        <div className="h-4 bg-gray-200 rounded w-1/2 mb-6"></div>
                        <div className="h-64 bg-gray-200 rounded"></div>
                    </div>
                </main>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-b from-background to-secondary/20">
            <DashboardHeader />

            <main className="container mx-auto px-4 py-6 md:px-6">
                <div className="space-y-6">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
                            Billing & Subscription
                        </h1>
                        <p className="text-muted-foreground mt-2">
                            Manage your subscription and billing details.
                        </p>
                    </div>

                    {/* Current Subscription Status */}
                    <Card className="border-primary/10 shadow-md">
                        <CardHeader className="pb-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                    <CreditCard className="h-6 w-6 text-primary" />
                                    <CardTitle className="text-xl">Current Plan</CardTitle>
                                </div>
                                <Badge
                                    variant={subscription.isActive ? "default" : "secondary"}
                                    className="text-sm"
                                >
                                    {subscription.isActive ? "Active" : "Inactive"}
                                </Badge>
                            </div>
                            <CardDescription className="mt-1">
                                You are currently on the {subscription.plan} plan
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="pt-0">
                            <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <Calendar className="h-4 w-4" />
                                <span>
                                    {subscription.plan === "FREE"
                                        ? `Expires at: ${formatTimestampToReadableDate(subscription.expiryDate)}`
                                        : `Next billing: ${formatTimestampToReadableDate(subscription.expiryDate)}`
                                    }
                                </span>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Upgrade your plan via chat card */}
                    <Card>
                        <CardContent className="!pt-6">
                            <h5 className="text-lg">{`Connect via chat to Upgrade.`}</h5>
                        </CardContent>
                    </Card>

                    {/* Choose Your Plan */}
                    <div className="space-y-4">
                        <h2 className="text-2xl font-semibold">Choose Your Plan</h2>
                        <div className="grid gap-6 md:grid-cols-3">
                            {plans.map((plan) => (
                                <Card
                                    key={plan.name}
                                    className={`relative border-primary/10 shadow-md ${plan.popular ? 'ring-2 ring-primary/20' : ''
                                        } ${subscription.plan === plan.name ? 'bg-primary/5 border-primary/30 ring-2 ring-primary/40' : ''
                                        }`}
                                >
                                    {plan.popular && subscription.plan !== plan.name && (
                                        <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                                            <Badge className="bg-primary text-primary-foreground px-3 py-1">
                                                Most Popular
                                            </Badge>
                                        </div>
                                    )}

                                    {subscription.plan === plan.name && (
                                        <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                                            <Badge className="bg-green-500 text-white px-3 py-1">
                                                Current Plan
                                            </Badge>
                                        </div>
                                    )}

                                    <CardHeader className="text-center pb-2">
                                        <div className="flex justify-center mb-2">
                                            {plan.icon}
                                        </div>
                                        <CardTitle className="text-xl">{plan.name}</CardTitle>
                                        <p className="text-sm text-muted-foreground mb-2">{plan.description}</p>
                                        <CardDescription className="text-2xl font-bold text-foreground">
                                            {plan.price}
                                            <span className="text-sm font-normal text-muted-foreground">
                                                {" "}{plan.interval}
                                            </span>
                                        </CardDescription>
                                    </CardHeader>

                                    <CardContent className="space-y-4">
                                        <div className="space-y-2">
                                            {plan.features.map((feature, index) => (
                                                <div key={index} className="flex items-center gap-2 text-sm">
                                                    <Check className="h-4 w-4 text-green-500 flex-shrink-0" />
                                                    <span>{feature}</span>
                                                </div>
                                            ))}
                                        </div>

                                        {/* <Button
                                            onClick={() => handleUpgradePlan(plan.name)}
                                            className={`w-full ${subscription.plan === plan.name
                                                ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                                                : plan.popular
                                                    ? 'bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90'
                                                    : ''
                                                }`}
                                            variant={plan.popular && subscription.plan !== plan.name ? 'default' : 'outline'}
                                            disabled={subscription.plan === plan.name}
                                        >
                                            {subscription.plan === plan.name ? 'Current Plan' : plan.name === 'FREE' ? 'Start Free Trial' : `Upgrade to ${plan.name}`}
                                        </Button> */}
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    </div>

                </div>
            </main>
        </div>
    );
}