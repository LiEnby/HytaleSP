# Aurora

This is a internal memory patcher for the Hytale Client;

on windows, hytale (well actually sentry ..)  will try to automatically load "Secur32.dll" 
the only function it imports from it is GetUserNameW which is used to send your windows username
as telemetry back to hytale servers, ( an anti-feature to begin with. )

simply stubbing that with a "return 0;" seems to work pretty well-

(and on linux you can just use ``LD_PRELOAD``, and macos ``DYLD_INSERT_LIBRARIES `` environment variables.)

this makes it very useful for getting our own code running in the context of the hytale process;

currently i am just using this to replace "account-data.hytale.com" and "session.hytale.com", etc; with localhost;
then- the launcher implements a "server emulator" for the authentication servers.

i use this method as i feel it's probably more resiliant to updates 
as it does not requiring any changes for new versions, unless something actually changes about the game;

but also more importantly to avoid shipping an unknown executable file and potentially breaking
itch.io's "Wharf" incremental patch system;

.. but also- this method could be easily used to make hacked clients or other more extensive modifications
currently hytale has no client-side anticheat of any kind; despite their website saying otherwise.


## Naming explaination

i was looking in procmon for anything loaded by the game, 
and i noticed it seemed to attempt to open "Aurora.dll" but it always failed;
curious about this i made my own dll named "Aurora.dll" and put it in the game folders--
which didn't seem to do anything-

then i found Secur32.dll, but at that point i had already made a project in VS2026.