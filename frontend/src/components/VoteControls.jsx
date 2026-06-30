import React from 'react';
import { BiSolidLike, BiSolidDislike } from "react-icons/bi";

export const VoteControls = ({ /* ... props */ }) => {

  return (
    <div className="flex items-center space-x-2 vote-controls">
      {/* Like Button */}
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation(); 
          handleVote('LIKE');
        }}
        disabled={isDisabled || isMutating}
        aria-pressed={currentVote === 'LIKE'}
        aria-label={`Like this ${targetType}. Current likes: ${likes}`}
        className={`flex items-center p-2 rounded transition-colors ${
          currentVote === 'LIKE' ? 'text-blue-600 bg-blue-50' : 'text-gray-500 hover:bg-gray-100'
        } disabled:opacity-50`}
      >
        <BiSolidLike aria-hidden="true" size={24} className="mr-1" />
        <span className="text-sm font-medium">{likes}</span>
      </button>

      {/* Dislike Button */}
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          handleVote('DISLIKE');
        }}
        disabled={isDisabled || isMutating}
        aria-pressed={currentVote === 'DISLIKE'}
        aria-label={`Dislike this ${targetType}. Current dislikes: ${dislikes}`}
        className={`flex items-center p-2 rounded transition-colors ${
          currentVote === 'DISLIKE' ? 'text-red-600 bg-red-50' : 'text-gray-500 hover:bg-gray-100'
        } disabled:opacity-50`}
      >
        <BiSolidDislike aria-hidden="true" size={24} className="mr-1" />
        <span className="text-sm font-medium">{dislikes}</span>
      </button>
    </div>
  );
};