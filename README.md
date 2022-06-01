# Back-end Coding Challenge

Alternative implementation of Q3 and Q4 when the actions list is not sorted.

## Implementation details

This implementation does **not** sort the actions list in before hand.

**Q1** and **Q2** stay the same.

For **Q3**, a map is created to store an ordered list of each user's actions (of any type). We populate the map by looping through the actions slice once. At the same time, we store a reference to each action of the specified type in another map, again organized by the action's user. This is the most complex part of the implementation, around O(#actions * (#actions/#users)) if the actions are equally distributed per user, or O(#actions^2) in the worst case where all actions are taken by the same user.

Next, we go through each reference we stored, and obtain what action is allocated next in the list for that same user. Because the references are pointers to list elements, who also have pointers to the "Next" and "Previous" elements in the list, this is done in constant time, so at the end the complexity is O(#specifiedTypeActions). 

Finally, we calculate the percentages for each action type.

Further benchmarking and data analysis should be done to clarify which solution is faster: the previous one, which would require sorting the actions list in before hand, or this one. 

For **Q4**, the solution has only changed slightly: we loop through the actions list to store the amount of direct referrals for each user, as well as who their referrer is (if any), thus creating a tree structure. At the same time, we store the ids of the users that have been referred, but have not taken a referral action themselves: they will be the leaves in our tree.

Then, we navigate the tree in reverse, starting from the leaves, and continuing to their parents, and their parents... so on. For each node, we add to its parent's referral index the node's own referral index, so indirect referrals are included in the final index.

Similarly as in the previous implementation, the complexity should be lineal, roughly of O(#actions) + O(#referrers + #referralTargets)

**Note**, referrals to oneself are not taken into account, otherwise the last loop would run infinitely
