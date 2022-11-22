> Work in progress

# hashsnap
Remove files ("trim") that have duplicates elsewhere.

By using space-savvy hash snapshots, *hashsnap* can perform deduplication across
multiple filesystems without having them all accessed at the same time.

In a nutshell:
- Run ```hashsnap create``` on some NAS
- Copy the generated ```.hsnap``` file on your localhost, rename it ```nas.hsnap```
- Run ```hashsnap trim nas.hsnap -delete``` to remove all local files that have duplicates
on the NAS

Generated .hsnap files are small: 6GB of results in a ~360k hashsnap file...

... and fast to generate: on a quadcore+SSD, it takes around 2min to hash said 6GB.

A Synology NAS hashes 100GB within 26min.

Duplication finding uses file size and SHA-1 hashing comparison, so don't expect security, 
but fair level of guarantee that only duplicates are found.

## Usage
    
    hashsnap help
    hashsnap <cmd> -help

Running on a NAS without screen/tmux and forgetting about it:

    nohup hashsnap... </dev/null >hashsnap.log 2>&1 &

Exploring easily an info result:

    hashsnap trim nas.hsnap | less -R

## TODO
- be able to resume a hashing, by leveraging file existence,
- test extensively,
- dedup without trimming (perhaps, as fdupes does it already quite well),
- cleanup the code.

## Inspiration
- [mathieuancelin/duplicates](https://github.com/mathieuancelin/duplicates/blob/master/duplicates.go)
- [SeeSpotRun gist](https://gist.github.com/SeeSpotRun/456b88424841d7ae735f)

https://goreleaser.com/
