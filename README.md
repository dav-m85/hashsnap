Work in progress

# hashsnap
Create a hash snapshot of a file tree, and perform deduplication using multiple snapshots.

## Usage
Say you have a directory `somedir`, and want to deduplicate it.

    cd somedir
    hashsnap create somedir.hsnap
    hashsnap dedup somedir.hsnap

Now, say you have a NAS where you have already stored a few files of somedir, and have
them removed of somedir

    # On the NAS
    hashsnap create nas.hsnap

    # In somedir
    hashsnap dedup somedir.hsnap --with nas.hsnap


## Bench
On my quadcore + SSD workstation, it takes ~2min to hash 6GB of data, resulting in a ~360k hashsnap file.

## TODO
- be able to resume a hashing, by leveraging file existence.
- be able to "check" a snapshot, for file existence.
- store path structure along with file, in order to make smart decisions when deleting.
- test extensively.
- BUG when a file is duplicated and --with is used
- CPU looks to be the bottleneck on create command. However, hard to not hash all the files, since we
want to compare with distant files. Find maybe a smarter way of hashing ? different algo ?

## Inspiration
- [mathieuancelin/duplicates](https://github.com/mathieuancelin/duplicates/blob/master/duplicates.go)
- [SeeSpotRun gist](https://gist.github.com/SeeSpotRun/456b88424841d7ae735f)
