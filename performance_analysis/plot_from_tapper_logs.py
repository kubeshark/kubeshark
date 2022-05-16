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


def append_sample(name: str, line: str, samples: typing.List[float]):
    pattern = name + r': ?(\d+(\.\d+)?)'
    maybe_sample = re.findall(pattern, line)
    if len(maybe_sample) == 0:
        return

    sample = float(maybe_sample[0][0])
    samples.append(sample)


def extract_samples(f: typing.IO) -> typing.Tuple[pd.Series, pd.Series, pd.Series]:
    cpu_samples = []
    rss_samples = []
    count_samples = []
    for line in f:
        append_sample('cpu', line, cpu_samples)
        append_sample('rss', line, rss_samples)
        append_sample('"packetsCount"', line, count_samples)

    cpu_samples = pd.Series(cpu_samples)
    rss_samples = pd.Series(rss_samples)
    count_samples = pd.Series(count_samples)

    return cpu_samples, rss_samples, count_samples


def plot(df: pd.DataFrame, title: str, xlabel: str, ylabel: str, group_pattern: typing.Optional[str]):
    if group_pattern:
        color = get_group_color(df.columns, group_pattern)
        df.plot(color=color, ax=ax)
    else:
        df.plot(cmap=COLORMAP, ax=ax)

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

    for ii, filename in enumerate(filenames):
        with open(filename, 'r') as f:
            cpu_samples, rss_samples, count_samples = extract_samples(f)

        cpu_samples.name = pathlib.Path(filename).name
        rss_samples.name = pathlib.Path(filename).name
        count_samples.name = pathlib.Path(filename).name

        cpu_samples_all_files.append(cpu_samples)
        rss_samples_all_files.append(rss_samples)
        count_samples_all_files.append(count_samples)

    cpu_samples_df = pd.concat(cpu_samples_all_files, axis=1)
    rss_samples_df = pd.concat(rss_samples_all_files, axis=1)
    count_samples_df = pd.concat(count_samples_all_files, axis=1)

    group_pattern = r'^\d+'

    ax = plt.subplot(3, 1, 1)
    plot(cpu_samples_df, 'cpu', '# sample', 'cpu (%)', group_pattern)

    ax = plt.subplot(3, 1, 2)
    plot((rss_samples_df / 1024 / 1024), 'rss', '# sample', 'mem (MB)', group_pattern)

    ax = plt.subplot(3, 1, 3)
    plot(count_samples_df, 'packetsCount', '# sample', 'packetsCount', group_pattern)

    plt.show()
