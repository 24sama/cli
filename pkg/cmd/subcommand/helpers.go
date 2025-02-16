package subcommand

import (
	"strings"

	client "github.com/OpenFunction/cli/pkg/client"
	fn "github.com/openfunction/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
)

func getFromFilenameOptions(cmd *cobra.Command, filenameOptions resource.FilenameOptions) ([]*fn.Function, error) {
	r := resource.NewLocalBuilder().
		WithScheme(client.Scheme, fn.GroupVersion).
		ContinueOnError().
		FilenameParam(false, &filenameOptions).
		Do()

	if err := r.Err(); err != nil {
		return nil, err
	}

	fns := make([]*fn.Function, 0)
	err := r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		obj1 := info.Object
		if obj1.GetObjectKind().GroupVersionKind().Group == fn.GroupVersion.Group {
			fn, ok := obj1.(*fn.Function)
			if ok {
				fns = append(fns, fn)
			}
		}

		return nil
	})

	return fns, err
}

func AddBuild(cmd *cobra.Command, builder *fn.BuildImpl) {
	var builderStr string
	builder.Builder = &builderStr
	cmd.Flags().StringVarP(builder.Builder, "builder", "", *builder.Builder, "Cloud Native Buildpacks builders")
	builder.SrcRepo = &fn.GitRepo{}
	builder.SrcRepo.Init()
	AddGitRepo(cmd, builder.SrcRepo)
}

func AddGitRepo(cmd *cobra.Command, gitRepo *fn.GitRepo) {
	cmd.Flags().StringVarP(&gitRepo.Url, "git-repo-url", "", gitRepo.Url, "Git url to clone")
	gitRepo.SourceSubPath = new(string)
	cmd.Flags().StringVarP(gitRepo.SourceSubPath, "git-repo-source-sub-path", "", *gitRepo.SourceSubPath, "A subpath within the `source` input where the source to build is located")
}

func AddServing(cmd *cobra.Command, serving *fn.ServingImpl) {
	var runtime string
	serving.Runtime = (*fn.Runtime)(&runtime)
	cmd.Flags().StringVarP(&runtime, "runtime", "", runtime, "Cloud Native Buildpacks builders")
}

func AddFilenameOptionFlags(cmd *cobra.Command, options *resource.FilenameOptions, usage string) {
	AddJsonFilenameFlag(cmd.Flags(), &options.Filenames, "Filename, directory, or URL to files "+usage)
	AddKustomizeFlag(cmd.Flags(), &options.Kustomize)
	cmd.Flags().BoolVarP(&options.Recursive, "recursive", "R", options.Recursive, "Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.")
}

func AddJsonFilenameFlag(flags *pflag.FlagSet, value *[]string, usage string) {
	flags.StringSliceVarP(value, "filename", "f", *value, usage)
	annotations := make([]string, 0, len(resource.FileExtensions))
	for _, ext := range resource.FileExtensions {
		annotations = append(annotations, strings.TrimLeft(ext, "."))
	}
	flags.SetAnnotation("filename", cobra.BashCompFilenameExt, annotations)
}

// AddKustomizeFlag adds kustomize flag to a command
func AddKustomizeFlag(flags *pflag.FlagSet, value *string) {
	flags.StringVarP(value, "kustomize", "k", *value, "Process the kustomization directory. This flag can't be used together with -f or -R.")
}

type PrintHandler interface {
	TableHandler(columns []metav1beta1.TableColumnDefinition, printFunc interface{}) error
}

// NewListFlags returns flags associated with humanreadable,
// template, and "name" printing, with default values set.
func NewListPrintFlags(printer func(h PrintHandler)) (*genericclioptions.PrintFlags, func(h PrintHandler)) {
	return genericclioptions.NewPrintFlags(""), printer
}
