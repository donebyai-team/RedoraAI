'use client'

import { Post, PostSettings } from "@doota/pb/doota/core/v1/post_pb";
import { useRouter } from "next/navigation";
import { useAppDispatch } from "@/store/hooks";
import toast from "react-hot-toast";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { setPost } from "@/store/PostCreation/PostCreationSlice";

/**
 * Hook to create a new post and update Redux state.
 * @returns { createPost } - function to create post
 */
export function useCreatePost() {
    const router = useRouter();
    const dispatch = useAppDispatch();
    const { portalClient } = useClientsContext();

    const createPost = async (
        postData: Omit<PostSettings, "$typeName">,
        redirect: boolean = true,
        setLoading: (val: boolean) => void
    ): Promise<Post | undefined> => {
        setLoading(true);

        try {
            const res = await portalClient.createPost(postData);
            dispatch(setPost(res));
            toast.success("Post created successfully!");

            if (redirect) {
                setTimeout(() => {
                    router.push("/post-creation-hub/editor");
                }, 500);
            }

            return res;
        } catch (err: any) {
            const message = err?.response?.data?.message || err.message || "Something went wrong";
            toast.error(message);
            return undefined;
        } finally {
            setLoading(false);
        }
    };

    return { createPost };
}
