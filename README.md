# OliGO

## Usage Guide

### 1- Generating Candidate Probes

Using `blockParse`, generate the candidates.

Example using a hypothetical `example.fasta` genomic file, outputting a file called `example_probes.fasta`:

	blockParse -target example.fasta -out example_probes.fasta
	
For further help on what is configurable, use the following command:

	blockParse -h

### 2- Aligning Candidates to a Target Genome

For this step, we will use [Bowtie2](https://github.com/BenLangmead/bowtie2) to align the probes to the target genome, retaining those that match sufficiently well.

Start by building a Bowtie index of your genomic file using `bowtie2-build`:

	bowtie2-build example.fasta example_index

Then align the generated probes to the genomic file:

	bowtie2 -x example_index -k 2 -f -U example_probes.fasta --score-min L,0,-1.5 -S aligned_probes.sam

### 3- Filtering Aligned Probes

Run `filterProbes` with the SAM output of the previous step.

Example using our hypothetical `aligned_probes.sam`, removing non-unique probes:

	filterProbes -target aligned_probes.sam -unique -out filtered_probes.fasta

For further help on what is configurable, use the following command:

	filterProbes -h

**Attention!** `filterProbes` will automatically filter out probes that have low MAPQ score. If you don't want this behaviour, use the flag `-minMAPQ 0`.
