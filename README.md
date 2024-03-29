> Work in progress

# hsnap
Remove files ("trim") that have duplicates elsewhere.

By using space-savvy hash snapshots, *hsnap* can perform deduplication across
multiple filesystems without having them all accessed at the same time.

In a nutshell:
- Run ```hsnap create``` on some NAS
- Copy the generated ```.hsnap``` file on your localhost, rename it ```nas.hsnap```
- Run ```hsnap trim nas.hsnap -delete``` to remove all local files that have duplicates
on the NAS

Generated .hsnap files are small: 6GB of results in a ~360k hsnap file...

... and fast to generate: on a quadcore+SSD, it takes around 2min to hash said 6GB.

A Synology NAS hashes 100GB within 26min.

Duplication finding uses file size and SHA-1 hashing comparison, so don't expect security, 
but fair level of guarantee that only duplicates are found.

## Usage
    
    hsnap help
    hsnap <cmd> -help

Running on a NAS without screen/tmux and forgetting about it:

    nohup hsnap... </dev/null >hsnap.log 2>&1 &

Exploring easily an info result:

    hsnap trim nas.hsnap | less -R

## TODO
- be able to resume a hashing, by leveraging file existence,
- test extensively,
- dedup without trimming (perhaps, as fdupes does it already quite well),
- cleanup the code.
- Buffer .hsnap while creating in memory...

## Inspiration
- [mathieuancelin/duplicates](https://github.com/mathieuancelin/duplicates/blob/master/duplicates.go)
- [SeeSpotRun gist](https://gist.github.com/SeeSpotRun/456b88424841d7ae735f)

https://goreleaser.com/
