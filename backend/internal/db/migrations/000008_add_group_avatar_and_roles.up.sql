ALTER TABLE groups ADD COLUMN avatar TEXT;

ALTER TABLE group_members ADD COLUMN role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member'));

UPDATE group_members
SET role = 'admin'
WHERE status = 'accepted'
  AND EXISTS (
    SELECT 1
    FROM groups g
    WHERE g.id = group_members.group_id
      AND g.creator_id = group_members.user_id
  );
