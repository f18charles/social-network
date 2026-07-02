import { useCallback, useEffect, useState } from "react";
import { useLocation, useParams, useSearchParams } from "react-router";
import Post from "../components/Post.jsx";
import Comment from "../components/Comment.jsx";
import CommentComposer from "../components/CommentComposer.jsx";
import { apiFetch } from "../utils/api.js";
import { logger } from "../utils/logger.js";
import "../styles/post-detail.css";

const REPLY_PAGE_SIZE = 10;

const getListData = (response) =>
  Array.isArray(response) ? response : response?.data || [];
const getListPagination = (response) =>
  response?.pagination || { has_more: false };

const incrementReplyCount = (comment) =>
  comment?.deleted
    ? comment
    : { ...comment, replies_count: (comment?.replies_count || 0) + 1 };

const appendReply = (comments, parentID, reply) =>
  comments.map((comment) => {
    if (comment.id === parentID) {
      return incrementReplyCount({
        ...comment,
        replies: [...(comment.replies || []), reply],
      });
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      const nextReplies = appendReply(comment.replies, parentID, reply);
      const changed = nextReplies.some(
        (item, index) => item !== comment.replies[index]
      );
      return changed
        ? incrementReplyCount({ ...comment, replies: nextReplies })
        : comment;
    }

    return comment;
  });

const mergeReplies = (comments, parentID, replies) =>
  comments.map((comment) => {
    if (comment.id === parentID) {
      const existing = new Set(
        (comment.replies || []).map((reply) => reply.id)
      );
      const nextReplies = [
        ...(comment.replies || []),
        ...replies.filter((reply) => !existing.has(reply.id)),
      ];
      return { ...comment, replies: nextReplies };
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      return {
        ...comment,
        replies: mergeReplies(comment.replies, parentID, replies),
      };
    }

    return comment;
  });

const ensureRootComment = (comments, rootComment) => {
  if (!rootComment || comments.some((comment) => comment.id === rootComment.id))
    return comments;
  return [...comments, { ...rootComment, replies: rootComment.replies || [] }];
};

const updateCommentVote = (comments, commentID, summary) =>
  comments.map((comment) => {
    if (comment.id === commentID) {
      return {
        ...comment,
        like_count: summary?.like_count ?? comment.like_count ?? 0,
        dislike_count: summary?.dislike_count ?? comment.dislike_count ?? 0,
        viewer_vote: summary?.viewer_vote || "none",
      };
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      return {
        ...comment,
        replies: updateCommentVote(comment.replies, commentID, summary),
      };
    }

    return comment;
  });


const replaceComment = (comments, updatedComment) =>
  comments.map((comment) => {
    if (comment.id === updatedComment?.id) {
      return { ...comment, ...updatedComment, replies: updatedComment.replies || comment.replies || [] };
    }

    if (Array.isArray(comment.replies) && comment.replies.length > 0) {
      return {
        ...comment,
        replies: replaceComment(comment.replies, updatedComment),
      };
    }

    return comment;
  });

const findComment = (comments, commentID) => {
  for (const comment of comments) {
    if (comment.id === commentID) return comment;
    const found = Array.isArray(comment.replies)
      ? findComment(comment.replies, commentID)
      : null;
    if (found) return found;
  }
  return null;
};

/**
 * PostDetail renders a single post and a folded database-backed comment thread.
 */
const PostDetail = () => {
  const { id } = useParams();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const focusedCommentQuery = searchParams.get("comment_id");
  const [post, setPost] = useState(location.state || null);
  const [comments, setComments] = useState([]);
  const [loadingReplies, setLoadingReplies] = useState({});
  const [loading, setLoading] = useState(!location.state);
  const [error, setError] = useState(null);
  const [focusedCommentId, setFocusedCommentId] = useState(focusedCommentQuery);

  const fetchRepliesPage = useCallback(async (commentID, offset = 0) => {
    const response = await apiFetch(
      `/api/comments/${commentID}/replies?limit=${REPLY_PAGE_SIZE}&offset=${offset}`
    );
    return {
      replies: getListData(response),
      pagination: getListPagination(response),
    };
  }, []);

  const revealCommentPath = useCallback(
    async (commentID, initialComments, isActive) => {
      try {
        const context = await apiFetch(`/api/comments/${commentID}/context`);
        const path = Array.isArray(context?.path) ? context.path : [];
        if (!path.length || !isActive()) return;

        let tree = ensureRootComment(initialComments, path[0]);
        for (let index = 0; index < path.length - 1; index += 1) {
          const parent = path[index];
          const child = path[index + 1];
          let offset = 0;
          let found = false;

          while (!found) {
            const { replies, pagination } = await fetchRepliesPage(
              parent.id,
              offset
            );
            tree = mergeReplies(tree, parent.id, replies);
            found = replies.some((reply) => reply.id === child.id);
            if (found || !pagination.has_more) break;
            offset += REPLY_PAGE_SIZE;
          }
        }

        if (!isActive()) return;
        setComments(tree);
        setFocusedCommentId(commentID);
        window.setTimeout(() => {
          document.getElementById(`comment-${commentID}`)?.scrollIntoView({
            behavior: "smooth",
            block: "center",
          });
        }, 0);
      } catch (err) {
        logger.error("Failed to reveal linked comment", err, {
          postId: id,
          commentId: commentID,
        });
      }
    },
    [fetchRepliesPage, id]
  );

  useEffect(() => {
    let active = true;

    async function loadPostDetail() {
      setLoading(true);
      setError(null);
      try {
        const [postData, commentData] = await Promise.all([
          apiFetch(`/api/posts/${id}`),
          apiFetch(`/api/posts/${id}/comments`),
        ]);
        if (!active) return;
        const topLevelComments = getListData(commentData);
        setPost(postData);
        setComments(topLevelComments);

        if (focusedCommentQuery) {
          revealCommentPath(
            focusedCommentQuery,
            topLevelComments,
            () => active
          );
        }
      } catch (err) {
        logger.error("Failed to load post detail", err, { postId: id });
        if (!active) return;
        setError("Unable to load post.");
      } finally {
        if (active) setLoading(false);
      }
    }

    if (id) loadPostDetail();

    return () => {
      active = false;
    };
  }, [id, focusedCommentQuery, revealCommentPath]);

  const loadReplies = async (commentID, offset = 0) => {
    setLoadingReplies((current) => ({ ...current, [commentID]: true }));
    try {
      const { replies } = await fetchRepliesPage(commentID, offset);
      setComments((current) => mergeReplies(current, commentID, replies));
    } catch (err) {
      logger.error("Failed to load comment replies", err, {
        commentId: commentID,
        offset,
      });
    } finally {
      setLoadingReplies((current) => ({ ...current, [commentID]: false }));
    }
  };

  const createComment = async (formData, parentCommentID = null) => {
    if (parentCommentID) formData.append("parent_comment_id", parentCommentID);
    const created = await apiFetch(`/api/posts/${id}/comments`, {
      method: "POST",
      body: formData,
    });

    if (parentCommentID) {
      setComments((current) => appendReply(current, parentCommentID, created));
    } else {
      setComments((current) => [...current, created]);
      setPost((current) =>
        current
          ? { ...current, comment_count: (current.comment_count || 0) + 1 }
          : current
      );
    }
  };

  const voteComment = async (commentID, vote) => {
    try {
      const target = findComment(comments, commentID);
      const summary =
        target?.viewer_vote === vote
          ? await apiFetch(`/api/comments/${commentID}/vote`, {
              method: "DELETE",
            })
          : await apiFetch(`/api/comments/${commentID}/vote`, {
              method: "PUT",
              body: { vote },
            });
      setComments((current) => updateCommentVote(current, commentID, summary));
    } catch (err) {
      logger.error("Failed to update comment vote", err, {
        commentId: commentID,
        vote,
      });
    }
  };

  const updateComment = (updatedComment) => {
    setComments((current) => replaceComment(current, updatedComment));
  };

  return (
    <div className="post-detail">
      {loading && <div>Loading post...</div>}
      {error && <div className="error">{error}</div>}
      {post && <Post post={post} onPostChange={setPost} />}
      <div className="comments card">
        <h3>Comments</h3>
        {post && !post.deleted ? (
          <CommentComposer onSubmit={(formData) => createComment(formData)} />
        ) : null}
        {comments.length === 0 && !loading ? <p>No comments yet.</p> : null}
        {comments.map((comment) => (
          <Comment
            comment={comment}
            key={comment.id}
            isPostDeleted={post?.deleted}
            onCreateReply={(parentID, formData) =>
              createComment(formData, parentID)
            }
            onUpdateComment={updateComment}
            onVote={voteComment}
            onLoadReplies={loadReplies}
            loadingReplies={loadingReplies}
            focusedCommentId={focusedCommentId}
          />
        ))}
      </div>
    </div>
  );
};

export default PostDetail;
