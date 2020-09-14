% Allow self-code-review in certain branches of this repo.
% See All-Projects rules.pl for details.

submit_rule(S) :-
	Gobot = user(5976),
	gerrit:default_submit(R),
	R =.. [submit | A],
	limited_self_review_ok(Gobot, A, B),
	S =.. [submit | B].

limited_self_review_ok(Gobot, In, Out) :- ok_for_self_review, !, Out = [label('Self-Review-OK', ok(Gobot)) | In].
limited_self_review_ok(Gobot, In, Out) :- Out = In.

% dev.go2go is an experimental prototype that won't be used for production use.
% The owners - gri and iant - are allowed to self-review changes there.
ok_for_self_review :- gerrit:change_branch('refs/heads/dev.go2go'), gerrit:commit_author(_, _, 'gri@golang.org').
ok_for_self_review :- gerrit:change_branch('refs/heads/dev.go2go'), gerrit:commit_author(_, _, 'iant@golang.org').
