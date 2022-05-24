import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import pathlib
import re
import sys
import typing

COLORMAP = plt.get_cmap('turbo')

# Extract cpu and rss samples from log files and plot them
# Input: List of log files
#
# example:
#   python plot_from_tapper_logs.py 01_no_pcap_01.log 99_normal_00.log
#
# The script assumes that the log file names start with a number (pattern '\d+')
# and groups based on this number. Files that start will the same number will be plotted with the same color.
# Change group_pattern to an empty string to disable this, or change to a regex of your liking.


def get_sample(name: str, line: str, default_value: float):
    pattern = name + r': ?(\d+(\.\d+)?)'
    maybe_sample = re.findall(pattern, line)
    if len(maybe_sample) == 0:
        return default_value

    sample = float(maybe_sample[0][0])
    return sample


def append_sample(name: str, line: str, samples: typing.List[float]):
    sample = get_sample(name, line, -1)

    if sample == -1:
        return

    samples.append(sample)


def extract_samples(f: typing.IO) -> typing.Tuple[pd.Series, pd.Series, pd.Series, pd.Series, pd.Series, pd.Series, pd.Series, pd.Series]:
    cpu_samples = []
    rss_samples = []
    count_samples = []
    matched_samples = []
    live_samples = []
    processed_samples = []
    heap_samples = []
    goroutines_samples = []
    for line in f:
        append_sample('cpu', line, cpu_samples)
        append_sample('rss', line, rss_samples)
        ignored_packets_count = get_sample('"ignoredPacketsCount"', line, -1)
        packets_count = get_sample('"packetsCount"', line, -1)
        if ignored_packets_count != -1 and packets_count != -1:
            count_samples.append(packets_count - ignored_packets_count)
        append_sample('"matchedPairs"', line, matched_samples)
        append_sample('"liveTcpStreams"', line, live_samples)
        append_sample('"processedBytes"', line, processed_samples)
        append_sample('mem', line, heap_samples)
        append_sample('goroutines', line, goroutines_samples)

    cpu_samples = pd.Series(cpu_samples)
    rss_samples = pd.Series(rss_samples)
    count_samples = pd.Series(count_samples)
    matched_samples = pd.Series(matched_samples)
    live_samples = pd.Series(live_samples)
    processed_samples = pd.Series(processed_samples)
    heap_samples = pd.Series(heap_samples)
    goroutines_samples = pd.Series(goroutines_samples)

    return cpu_samples, rss_samples, count_samples, matched_samples, live_samples, processed_samples, heap_samples, goroutines_samples


def plot(ax, df: pd.DataFrame, title: str, xlabel: str, ylabel: str, group_pattern: typing.Optional[str]):
    if group_pattern:
        color = get_group_color(df.columns, group_pattern)
        df.plot(color=color, ax=ax)
    else:
        df.plot(cmap=COLORMAP, ax=ax)

    ax.ticklabel_format(style='plain')
    plt.title(title)
    plt.legend()
    plt.xlabel(xlabel)
    plt.ylabel(ylabel)


def get_group_color(names, pattern):
    props = [int(re.findall(pattern, pathlib.Path(name).name)[0]) for name in names]
    key = dict(zip(sorted(list(set(props))), range(len(set(props)))))
    n_colors = len(key)
    color_options = plt.get_cmap('jet')(np.linspace(0, 1, n_colors))
    groups = [key[prop] for prop in props]
    color = color_options[groups]  # type: ignore
    return color


if __name__ == '__main__':
    filenames = sys.argv[1:]

    cpu_samples_all_files = []
    rss_samples_all_files = []
    count_samples_all_files = []
    matched_samples_all_files = []
    live_samples_all_files = []
    processed_samples_all_files = []
    heap_samples_all_files = []
    goroutines_samples_all_files = []

    for ii, filename in enumerate(filenames):
        print("Analyzing {}".format(filename))
        with open(filename, 'r') as f:
            cpu_samples, rss_samples, count_samples, matched_samples, live_samples, processed_samples, heap_samples, goroutines_samples = extract_samples(f)

        cpu_samples.name = pathlib.Path(filename).name
        rss_samples.name = pathlib.Path(filename).name
        count_samples.name = pathlib.Path(filename).name
        matched_samples.name = pathlib.Path(filename).name
        live_samples.name = pathlib.Path(filename).name
        processed_samples.name = pathlib.Path(filename).name
        heap_samples.name = pathlib.Path(filename).name
        goroutines_samples.name = pathlib.Path(filename).name

        cpu_samples_all_files.append(cpu_samples)
        rss_samples_all_files.append(rss_samples)
        count_samples_all_files.append(count_samples)
        matched_samples_all_files.append(matched_samples)
        live_samples_all_files.append(live_samples)
        processed_samples_all_files.append(processed_samples)
        heap_samples_all_files.append(heap_samples)
        goroutines_samples_all_files.append(goroutines_samples)

    cpu_samples_df = pd.concat(cpu_samples_all_files, axis=1)
    rss_samples_df = pd.concat(rss_samples_all_files, axis=1)
    count_samples_df = pd.concat(count_samples_all_files, axis=1)
    matched_samples_df = pd.concat(matched_samples_all_files, axis=1)
    live_samples_df = pd.concat(live_samples_all_files, axis=1)
    processed_samples_df = pd.concat(processed_samples_all_files, axis=1)
    heap_samples_df = pd.concat(heap_samples_all_files, axis=1)
    goroutines_samples_df = pd.concat(goroutines_samples_all_files, axis=1)

    group_pattern = r'^\d+'

    cpu_plot = plt.subplot(8, 2, 1)
    plot(cpu_plot, cpu_samples_df, 'cpu', '', 'cpu (%)', group_pattern)
    cpu_plot.legend().remove()

    mem_plot = plt.subplot(8, 2, 2)
    plot(mem_plot, (rss_samples_df / 1024 / 1024), 'rss', '', 'mem (mega)', group_pattern)
    mem_plot.legend(loc='center left', bbox_to_anchor=(1, 0.5))

    packets_plot = plt.subplot(8, 2, 3)
    plot(packets_plot, count_samples_df, 'packetsCount', '', 'packetsCount', group_pattern)
    packets_plot.legend().remove()

    matched_plot = plt.subplot(8, 2, 4)
    plot(matched_plot, matched_samples_df, 'matchedCount', '', 'matchedCount', group_pattern)
    matched_plot.legend().remove()

    live_plot = plt.subplot(8, 2, 5)
    plot(live_plot, live_samples_df, 'liveStreamsCount', '', 'liveStreamsCount', group_pattern)
    live_plot.legend().remove()

    processed_plot = plt.subplot(8, 2, 6)
    plot(processed_plot, (processed_samples_df / 1024 / 1024), 'processedBytes', '', 'bytes (mega)', group_pattern)
    processed_plot.legend().remove()

    heap_plot = plt.subplot(8, 2, 7)
    plot(heap_plot, (heap_samples_df / 1024 / 1024), 'heap', '', 'heap (mega)', group_pattern)
    heap_plot.legend().remove()

    goroutines_plot = plt.subplot(8, 2, 8)
    plot(goroutines_plot, goroutines_samples_df, 'goroutines', '', 'goroutines', group_pattern)
    goroutines_plot.legend().remove()

    fig = plt.gcf()
    fig.set_size_inches(20, 18)

    print('Saving graph to graph.png')
    plt.savefig('graph.png', bbox_inches='tight')
    