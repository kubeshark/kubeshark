import matplotlib.pyplot as plt
import pandas as pd
import pathlib
import re
import sys
import typing

COLORMAP = plt.get_cmap('turbo')

# Extract cpu and rss samples from log files and plot them
# Input: List of log files

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

def plot(df: pd.DataFrame, title: str, xlabel: str, ylabel: str):
    df.plot(cmap=COLORMAP, ax=ax)
    plt.title(title)
    plt.legend()
    plt.xlabel(xlabel)
    plt.ylabel(ylabel)

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

    ax = plt.subplot(3, 1, 1)
    plot(cpu_samples_df, 'cpu', '# sample', 'cpu (%)')

    ax = plt.subplot(3, 1, 2)
    plot(rss_samples_df, 'rss', '# sample', 'mem (MB)')

    ax = plt.subplot(3, 1, 3)
    plot(count_samples_df, 'packetsCount', '# sample', 'packetsCount')

    plt.show()
