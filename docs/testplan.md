# rough notes and plans for ongoing QA on kpcli 2

This is used instead of the wiki because it's easier for me :D

# issues:
  1. db created in app is not opening in keepass (solved, wasn't passing -kpversion 1 at the cl) [X]
  1. notes field not rendering well in show, multiple lines are getting corrupted, need to ellipse at newlines [X]
goal: to test that the cli tool is interoperable with the public app
  1. `pwd` is rendering with a `<nil>` at the end in nested groups [X]
  1. *important* nested subgroups created on the cli are getting orphaned from the root, may relate to pathing problem above (could be two different bugs in kpv1 and kpv2)
	1. need to make db init logic live in per-implementation libs
  1. v2 tests aren't passing without writing twice to stdin, makes no sense
  1. need kpversion to be tweakable by tests
  1. need to reinstate metadata creation
  1. Should there be a separate entry stcut in common? no idea what i was thinking at the time
  1. need `info` command to print info about the db
  1. `show` on v2 no longer shows metadata
   1. as part of this fix, make a "short" and "long" version of show so that you can view all the bullshit at once, if you so desire, or you can just see the key parts.  An alternative would be to use "select", but this would be easier
  1. make sure retrieving binaries works, make it work if not
  1. would love to see if protected notes work
  1. cannot add extended fields on cli
  1. v2 is always created with an empty attachment field
  1. `select` doesn't use the same format as `show`
  1. `select` should allow setting default fields
  1. need to allow creation of custom fields on the cli
  1. look in to in-memory protection
  1. CLI's `new` doesn't even ask for all default fields, needs option to add more, as noted above
  1. CLI not saving in subgroup with no error being thrown
  1. nice to have: saveas should allow file system autocompletion
  1. allow creating DB when a non existent DB is specified on cli
  1. initial creation is throwing errors because the database allegedly exists, bitching about lack of backup which isn't helpful, second creation attempt works. lol.
  1. `save` doesn't allow for key creation - we really need full key management
  1. password failures are still dumping a lot of bullshit
  1. no tests working with the front end in v2 mode, may just be something to note
# core test plan:
1. Run each operation by doing it in one tool, verifying that it appears identically in the app, verify that both apps see the changes
  1. Create Database - brand new from scratch
  1. Create entry - fill out one field of each type in a new entry (not using sample crap)
    1. types:
      1. searchable
      1. protected
      1. readonly
      1. all of the previous 3 together
      1. string
      1. longstring
      1. binary
  1. Create group
  1. Create entry in subgroup
  1. Read   entry - select each field individually and then one at a time
  1. Update entry - update one field of each type
  1. Delete entry - create in first app, delete in second

# Test runs
1. app -> cli
  1. Create Database - brand new from scratch [X]
  1. Create entry - fill out one field of each type in a new entry (not using sample crap) [X]
    1. types:
      1. searchable
      1. protected
      1. readonly
      1. all of the previous 3 together
      1. string
      1. longstring
      1. binary
  1. Create group [X]
  1. Create entry in subgroup [X]
1. cli -> app
  1. Create Database - brand new from scratch
  1. Create entry - fill out one field of each type in a new entry (not using sample crap)
    1. types:
      1. searchable
      1. protected
      1. readonly
      1. all of the previous 3 together
      1. string
      1. longstring
      1. binary
  1. Create group
  1. Create entry in subgroup
1. cli -> cli
