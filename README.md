This is a simple discord verification bot

Example config.yml:

```yaml
# bot token
token: "xxxx"

# the role that should be set once a member adds the thumb up reaction
verify_role_id: 0000000000000000000

# channel that holds the message to trigger the verification
verify_channel: "0000000000000000001"
```

Create a verify text channel. Post a message and add a thumb up reaction.
Create a "Verified" role that is allowed to see the other channels except the verify one.
The verified role also shouldn't be able to see the verify text channel. Or members get confused if the channel is still available after verification.

If a member add the thumb up reaction they will get the "Verified" role assiged.
That's really all the bot does. Setting a role once a reaction is posted.