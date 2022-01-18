import intersection from 'lodash/intersection';

export function arrayElementsComparetion(array1,array2){
    if(array1.length !== array2.length){
        return false;
    }
    return (intersection(array1,array2).length === array1.length);
}