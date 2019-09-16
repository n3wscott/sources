# XKCD Bot

This is a demo of polling for events using a CURL command as a sink. This is
made possible by sidecar injection. The source for this example has zero code --
it simply curls an `xkcd.com` endpoint and forwards the data to the sink (which
is the injected sidecar).

## To run

1. Make sure you have installed kubectl, ko, and the `config` at this
   repository's root.
1. Set up the default eventing trigger in your cluster.
1. Create a GChat webhook by opening a GChat room, clicking the room name at the
   top, and clicking "Configure webhooks". Add a new webhook and copy the link.
1. Encode the webhook URL as base64 and put it in `config/xkcd-secrets.yaml`
   where appropriate.

   You might use `echo -n 'example.com/?token=1234567890987654321' | base64` to
   encode your text properly.
1. Run `ko apply -f config/`.

Every 30 minutes, you will get a message about the latest XKCD comic.
