import matplotlib.pyplot as plt
import pandas as pd
import pathlib
import re
import sys
import typing


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

if __name__ == '__main__':
    filenames = sys.argv[1:]

    fig = plt.figure(1)

    for ii, filename in enumerate(filenames):
        with open(filename, 'r') as f:
            cpu_samples, rss_samples, count_samples = extract_samples(f)

        cpu_samples.name = pathlib.Path(filename).name
        rss_samples.name = pathlib.Path(filename).name
        count_samples.name = pathlib.Path(filename).name

        plt.subplot(3, 1, 1)
        cpu_samples.plot()

        plt.subplot(3, 1, 2)
        (rss_samples / 1024 / 1024).plot()

        plt.subplot(3, 1, 3)
        count_samples.plot()

    plt.subplot(3, 1, 1)
    plt.legend()
    plt.title('cpu')
    plt.xlabel('# sample')
    plt.ylabel('cpu (%)')

    plt.subplot(3, 1, 2)
    plt.title('rss')
    plt.legend()
    plt.xlabel('# sample')
    plt.ylabel('mem (MB)')

    plt.subplot(3, 1, 3)
    plt.title('packetsCount')
    plt.legend()
    plt.xlabel('# sample')
    plt.ylabel('packetsCount')

    plt.show()
