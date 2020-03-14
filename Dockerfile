FROM scratch
COPY leaderz /usr/bin/leaderz
ENTRYPOINT ["leaderz"]
