# OliGO

## Usage Guide

### 1- Generating Candidate Probes

Using `blockParse`, generate the candidates.

Example using a hypothetical `example.fasta` genomic file, outputting a file called `example_probes.fasta`:

	blockParse -target example.fasta -out example_probes.fasta
	
For further help on what is configurable, use the following command:

	blockParse -h

### 2- Aligning Candidates to a Target Genome

For this step, you can use the aligner of your choice (as long as the output is a `.sam` file). For this example, we will be showing how to do it with [Bowtie2](https://github.com/BenLangmead/bowtie2).

Start by building a Bowtie index of your genomic file using `bowtie2-build`:

	bowtie2-build example.fasta example_index

Then align the generated probes to the genomic file:

	bowtie2 -x example_index -a -f example_probes.fasta -S aligned_probes.sam

**Attention!** It's important that the output is formatted as a SAM file or else it won't be possible to run the filter.

### 3- Filtering Aligned Probes

Run `filterProbes` with the SAM output of the previous step.

Example using our hypothetical `aligned_probes.sam`, removing non-unique probes:

	filterProbes -target aligned_probes.sam -unique -out filtered_probes.fasta

For further help on what is configurable, use the following command:

	filterProbes -h

**Attention!** `filterProbes` will automatically filter out probes that have low MAPQ score. If you don't want this behaviour, use the flag `-minMAPQ 0`.
