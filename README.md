> Work in progress

# hashsnap
Deduplicate files using snapshots, possibly from separate storage.

Usual approach for deduplicating files is "move all files on one storage, and run dedup there". What happens if no storage is big enough to hold all dedup files? Use snapshots.

Hashsnap generates hashed snapshots, which are space-savvy files that can be used to dedup on other storages without having all the data on it.

## Usage
Say you have a directory `somedir`, and want to deduplicate it.

    cd somedir
    # Creates an ".hsnap" file in somedir
    hashsnap create

    # State of hsnap in current directory
    hashsnap info [./some/dir/inside/somedir]

Now, say you have a NAS where you have already stored a few files of somedir, and have
them removed of somedir

    # On the NAS
    hashsnap create nas.hsnap

    # In somedir
    hashsnap trim nas.hsnap

Also deduplication works with:

    hashsnap dedup --keep ./somedir (will stop running on first undecided pair)

Running on a NAS without screen/tmux and forgetting about it:

    nohup hashsnap... </dev/null >hashsnap.log 2>&1 &

## Bench
On my quadcore + SSD workstation, it takes around 2min to hash 6GB of data, resulting in a ~360k hashsnap file.

## TODO
- be able to resume a hashing, by leveraging file existence.
- be able to "check" a snapshot, for file existence.
- store path structure along with file, in order to make smart decisions when deleting.
- test extensively.
- BUG when a file is duplicated and --with is used
- CPU looks to be the bottleneck on create command. However, hard to not hash all the files, since we
want to compare with distant files. Find maybe a smarter way of hashing ? different algo ?

scoring de similitude des emplacements de merge
nb de fichier identique

et swap file pour l'dit in place des fichiers hsnap

## Inspiration
- [mathieuancelin/duplicates](https://github.com/mathieuancelin/duplicates/blob/master/duplicates.go)
- [SeeSpotRun gist](https://gist.github.com/SeeSpotRun/456b88424841d7ae735f)
